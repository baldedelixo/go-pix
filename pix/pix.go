// Package pix generates Brazilian Pix Copy and Paste or QR Codes
//
// As a simple example:
//
// 	options := pix.Options{
// 		Name: "Jonnas Fonini",
// 		Key: "jonnasfonini@gmail.com",
// 		City: "Marau",
// 		Amount: 20.67,
// 		Description: "Invoice #4",
// 	}
//
// 	copyPaste, err := pix.Pix(options)
//
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	fmt.Println(copyPaste) // will output: "00020126580014BR.GOV.BCB.PIX0122jonnasfonini@gmail.com0210Invoice #4520400005303986540520.675802BR5913Jonnas Fonini6005Marau62410503***50300017BR.GOV.BCB.BRCODE01051.0.06304CF13"
//
package pix

import (
	"errors"
	"fmt"
	"github.com/r10r/crc16"
	"github.com/skip2/go-qrcode"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

// PixOptions is a configuration struct.
type PixOptions struct {
	// Pix Key (CPF/CNPJ, Email, Cellphone or Random Key)
	Key string
	// Receiver name
	Name string
	// Receiver city
	City string
	// Transaction amount
	Amount float64
	// Transaction description
	Description string
}

// QRCodeOptions is a configuration struct.
type QRCodeOptions struct {
	// QR Code content
	Content string
	// Default: 256
	Size int
}

type intMap map[int]interface{}

// Pix generates a Copy and Paste Pix code
func Pix(options PixOptions) (string, error) {
	if err := validateData(options); err != nil {
		return "", err
	}

	data := buildDataMap(options)

	str := parseData(data)

	// Add the CRC at the end
	str += "6304"

	crc, err := calculateCRC16(str)

	if err != nil {
		return "", err
	}

	str += crc

	return str, nil
}

// QRCode returns a graphical representation of the Copy and Paste code in a QR Code form.
func QRCode(options QRCodeOptions) ([]byte, error) {
	if options.Size == 0 {
		options.Size = 256
	}

	bytes, err := qrcode.Encode(options.Content, qrcode.Medium, options.Size)

	return bytes, err
}

func validateData(options PixOptions) error {
	if options.Key == "" {
		return errors.New("key must not be empty")
	}

	if utf8.RuneCountInString(options.Name) > 25 {
		return errors.New("name must be at least 25 characters long")
	}

	if utf8.RuneCountInString(options.City) > 15 {
		return errors.New("city must be at least 15 characters long")
	}

	return nil
}

func buildDataMap(options PixOptions) intMap {
	data := make(intMap)

	// Payload Format Indicator
	data[0] = "01"

	// Merchant Account Information
	data[26] = intMap{0: "BR.GOV.BCB.PIX", 1: options.Key, 2: options.Description}

	// Merchant Category Code
	data[52] = "0000"

	// Transaction Currency - Brazilian Real - ISO4217
	data[53] = "986"

	// Transaction Amount
	data[54] = options.Amount

	// Country Code - ISO3166-1 alpha 2
	data[58] = "BR"

	// Merchant Name. 25 characters maximum
	if options.Name != "" {
		data[59] = options.Name
	}

	// Merchant City. 15 characters maximum
	if options.City != "" {
		data[60] = options.City
	}

	// Transaction ID
	data[62] = intMap{5: "***", 50: intMap{0: "BR.GOV.BCB.BRCODE", 1: "1.0.0"}}

	return data
}

func parseData(data intMap) string {
	var str string

	keys := sortKeys(data)

	for _, k := range keys {
		v := reflect.ValueOf(data[k])

		switch v.Kind() {
		case reflect.String:
			str += leadingZeroIfLenSmallerThan10(strconv.Itoa(k)) + charCount(data[k].(string)) + data[k].(string)
		case reflect.Float64:
			value := strconv.FormatFloat(v.Float(), 'f', 2, 64)

			str += leadingZeroIfLenSmallerThan10(strconv.Itoa(k)) + charCount(value) + value
		case reflect.Map:
			// If the element is another map, do a recursive call
			content := parseData(data[k].(intMap))

			str += leadingZeroIfLenSmallerThan10(strconv.Itoa(k)) + charCount(content) + content
		}
	}

	return str
}

func sortKeys(data intMap) []int {
	keys := make([]int, len(data))
	i := 0

	for k := range data {
		keys[i] = k
		i++
	}

	sort.Ints(keys)

	return keys
}

func calculateCRC16(str string) (string, error) {
	table := crc16.MakeTable(crc16.CRC16_CCITT_FALSE)

	h := crc16.New(table)
	_, err := h.Write([]byte(str))

	if err != nil {
		return "", err
	}

	return strings.ToUpper(fmt.Sprintf("%x", h.Sum16())), nil
}

func leadingZeroIfLenSmallerThan10(str string) string {
	return left(str, 2, "0")
}

func charCount(str string) string {
	str = strconv.Itoa(len(str))

	return leadingZeroIfLenSmallerThan10(str)
}

func left(str string, length int, pad string) string {
	return times(pad, length-len(str)) + str
}

func times(str string, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(str, n)
}