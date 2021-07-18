package vad

import (
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/pandafw/pango/str"
)

// IsFileName is illegal file name
func IsFileName(s string) bool {
	return !str.ContainsAny(s, `\/:*?"<>|`)
}

// IsEmail checks if the string is an email.
func IsEmail(s string) bool {
	return rxEmail.MatchString(s)
}

// IsExistingEmail checks if the string is an email of existing domain
func IsExistingEmail(email string) bool {
	if len(email) < 6 || len(email) > 254 {
		return false
	}
	at := str.LastIndex(email, "@")
	if at <= 0 || at > len(email)-3 {
		return false
	}
	user := email[:at]
	host := email[at+1:]
	if len(user) > 64 {
		return false
	}

	if userDotRegexp.MatchString(user) || !userRegexp.MatchString(user) || !hostRegexp.MatchString(host) {
		return false
	}
	if _, err := net.LookupMX(host); err != nil {
		if _, err := net.LookupIP(host); err != nil {
			return false
		}
	}

	return true
}

const maxURLRuneCount = 2083
const minURLRuneCount = 3

// IsURL checks if the string is an URL.
func IsURL(s string) bool {
	if s == "" || utf8.RuneCountInString(s) >= maxURLRuneCount || len(s) <= minURLRuneCount || str.HasPrefix(s, ".") {
		return false
	}
	strTemp := s
	if str.Contains(s, ":") && !str.Contains(s, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + s
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return false
	}
	if str.HasPrefix(u.Host, ".") {
		return false
	}
	if u.Host == "" && (u.Path != "" && !str.Contains(u.Path, ".")) {
		return false
	}
	return rxURL.MatchString(s)
}

// IsRequestURL checks if the string rawurl, assuming
// it was received in an HTTP request, is a valid
// URL confirm to RFC 3986
func IsRequestURL(rawurl string) bool {
	url, err := url.ParseRequestURI(rawurl)
	if err != nil {
		return false //Couldn't even parse the rawurl
	}
	if len(url.Scheme) == 0 {
		return false //No Scheme found
	}
	return true
}

// IsRequestURI checks if the string rawurl, assuming
// it was received in an HTTP request, is an
// absolute URI or an absolute path.
func IsRequestURI(rawurl string) bool {
	_, err := url.ParseRequestURI(rawurl)
	return err == nil
}

// IsAlpha checks if the string contains only letters (a-zA-Z).
func IsAlpha(s string) bool {
	if s == "" {
		return false
	}

	l := len(s)
	for i := 0; i < l; i++ {
		b := s[i]
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')) {
			return false
		}
	}
	return true
}

//IsUTFLetter checks if the string contains only unicode letter characters.
//Similar to IsAlpha but for all languages.
func IsUTFLetter(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true

}

// IsAlphaNumeric checks if the string contains only letters and numbers.
func IsAlphaNumeric(s string) bool {
	if s == "" {
		return false
	}

	l := len(s)
	for i := 0; i < l; i++ {
		b := s[i]
		if !((b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')) {
			return false
		}
	}
	return true
}

// IsUTFLetterNumeric checks if the string contains only unicode letters and numbers.
func IsUTFLetterNumeric(s string) bool {
	if s == "" {
		return false
	}

	for _, c := range s {
		if !unicode.IsLetter(c) && !unicode.IsNumber(c) { //letters && numbers are ok
			return false
		}
	}
	return true

}

// IsNumeric checks if the string contains only numbers.
func IsNumeric(s string) bool {
	if s == "" {
		return false
	}

	l := len(s)
	for i := 0; i < l; i++ {
		b := s[i]
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// IsUTFNumeric checks if the string contains only unicode numbers of any kind.
// Numbers can be 0-9 but also Fractions ¾,Roman Ⅸ and Hangzhou 〩.
func IsUTFNumeric(s string) bool {
	if s == "" {
		return false
	}

	if str.IndexAny(s, "+-") > 0 {
		return false
	}
	if len(s) > 1 {
		s = str.TrimPrefix(s, "-")
		s = str.TrimPrefix(s, "+")
	}
	for _, c := range s {
		if !unicode.IsNumber(c) { //numbers && minus sign are ok
			return false
		}
	}
	return true

}

// IsUTFDigit checks if the string contains only unicode radix-10 decimal digits.
func IsUTFDigit(s string) bool {
	if s == "" {
		return false
	}

	if str.IndexAny(s, "+-") > 0 {
		return false
	}
	if len(s) > 1 {
		s = str.TrimPrefix(s, "-")
		s = str.TrimPrefix(s, "+")
	}
	for _, c := range s {
		if !unicode.IsDigit(c) { //digits && minus sign are ok
			return false
		}
	}
	return true

}

// IsHexadecimal checks if the string is a hexadecimal number.
func IsHexadecimal(s string) bool {
	if s == "" {
		return false
	}

	l := len(s)
	for i := 0; i < l; i++ {
		b := s[i]
		if !((b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F') || (b >= '0' && b <= '9')) {
			return false
		}
	}
	return true
}

// IsHexcolor checks if the string is a hexadecimal color.
func IsHexcolor(s string) bool {
	return rxHexcolor.MatchString(s)
}

// IsRGBcolor checks if the string is a valid RGB color in form rgb(RRR, GGG, BBB).
func IsRGBcolor(s string) bool {
	return rxRGBcolor.MatchString(s)
}

// IsLowerCase checks if the string is lowercase.
func IsLowerCase(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if unicode.IsUpper(r) {
			return false
		}
	}
	return true
}

// HasLowerCase checks if the string contains at least 1 lowercase.
func HasLowerCase(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

// IsUpperCase checks if the string is uppercase.
func IsUpperCase(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if unicode.IsLower(r) {
			return false
		}
	}
	return true
}

// HasUpperCase checks if the string contains as least 1 uppercase.
func HasUpperCase(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

// IsInt checks if the string is an integer.
func IsInt(s string) bool {
	if s == "" {
		return false
	}
	return rxInt.MatchString(s)
}

// IsFloat checks if the string is a float.
func IsFloat(s string) bool {
	return s != "" && rxFloat.MatchString(s)
}

// IsEmpty checks if the string is null.
func IsEmpty(s string) bool {
	return len(s) == 0
}

// IsNotEmpty checks if the string is not null.
func IsNotEmpty(s string) bool {
	return len(s) > 0
}

// IsWhitespace checks the string only contains whitespace
func IsWhitespace(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if !unicode.IsSpace(r) {
			return false
		}
	}
	return true
}

// HasWhitespace checks if the string contains any whitespace
func HasWhitespace(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if unicode.IsSpace(r) {
			return true
		}
	}
	return false
}

// IsByteLength checks if the string's length (in bytes) falls in a range.
func IsByteLength(s string, min, max int) bool {
	return len(s) >= min && len(s) <= max
}

// IsUUIDv3 checks if the string is a UUID version 3.
func IsUUIDv3(s string) bool {
	return rxUUID3.MatchString(s)
}

// IsUUIDv4 checks if the string is a UUID version 4.
func IsUUIDv4(s string) bool {
	return rxUUID4.MatchString(s)
}

// IsUUIDv5 checks if the string is a UUID version 5.
func IsUUIDv5(s string) bool {
	return rxUUID5.MatchString(s)
}

// IsUUID checks if the string is a UUID (version 3, 4 or 5).
func IsUUID(s string) bool {
	return rxUUID.MatchString(s)
}

// IsCreditCard checks if the string is a credit card.
func IsCreditCard(s string) bool {
	sanitized := str.RemoveAny(s, " -")
	if !rxCreditCard.MatchString(sanitized) {
		return false
	}

	var sum int64
	var digit string
	var tmpNum int64
	var shouldDouble bool
	for i := len(sanitized) - 1; i >= 0; i-- {
		digit = sanitized[i:(i + 1)]
		tmpNum, _ = strconv.ParseInt(digit, 0, 64)
		if shouldDouble {
			tmpNum *= 2
			if tmpNum >= 10 {
				sum += (tmpNum % 10) + 1
			} else {
				sum += tmpNum
			}
		} else {
			sum += tmpNum
		}
		shouldDouble = !shouldDouble
	}

	return sum%10 == 0
}

// IsISBN10 checks if the string is an ISBN version 10.
func IsISBN10(s string) bool {
	return IsISBN(s, 10)
}

// IsISBN13 checks if the string is an ISBN version 13.
func IsISBN13(s string) bool {
	return IsISBN(s, 13)
}

// IsISBN checks if the string is an ISBN (version 10 or 13).
// If version value is not equal to 10 or 13, it will be checks both variants.
func IsISBN(s string, version int) bool {
	sanitized := str.RemoveAny(s, " -")
	var checksum int32
	var i int32
	if version == 10 {
		if !rxISBN10.MatchString(sanitized) {
			return false
		}
		for i = 0; i < 9; i++ {
			checksum += (i + 1) * int32(sanitized[i]-'0')
		}
		if sanitized[9] == 'X' {
			checksum += 10 * 10
		} else {
			checksum += 10 * int32(sanitized[9]-'0')
		}
		if checksum%11 == 0 {
			return true
		}
		return false
	} else if version == 13 {
		if !rxISBN13.MatchString(sanitized) {
			return false
		}
		factor := []int32{1, 3}
		for i = 0; i < 12; i++ {
			checksum += factor[i%2] * int32(sanitized[i]-'0')
		}
		return (int32(sanitized[12]-'0'))-((10-(checksum%10))%10) == 0
	}
	return IsISBN(s, 10) || IsISBN(s, 13)
}

// IsJSON checks if the string is valid JSON (note: uses json.Unmarshal).
func IsJSON(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}

// IsMultibyte checks if the string contains one or more multibyte chars.
func IsMultibyte(s string) bool {
	if s == "" {
		return false
	}

	l := len(s)
	for i := 0; i < l; i++ {
		if s[i] > unicode.MaxASCII {
			return true
		}
	}
	return false
}

// IsASCII checks if the string contains ASCII chars only.
func IsASCII(s string) bool {
	return str.IsASCII(s)
}

// IsPrintableASCII checks if the string contains printable ASCII chars only.
func IsPrintableASCII(s string) bool {
	return str.IsPrintableASCII(s)
}

// IsFullWidth checks if the string contains any full-width chars.
func IsFullWidth(s string) bool {
	if s == "" {
		return false
	}

	return rxFullWidth.MatchString(s)
}

// IsHalfWidth checks if the string contains any half-width chars.
func IsHalfWidth(s string) bool {
	if s == "" {
		return false
	}
	return rxHalfWidth.MatchString(s)
}

// IsVariableWidth checks if the string contains a mixture of full and half-width chars.
func IsVariableWidth(s string) bool {
	if s == "" {
		return false
	}
	return rxHalfWidth.MatchString(s) && rxFullWidth.MatchString(s)
}

// IsBase64 checks if a string is base64 encoded.
func IsBase64(s string) bool {
	return rxBase64.MatchString(s)
}

// IsDataURI checks if a string is base64 encoded data URI such as an image
func IsDataURI(s string) bool {
	dataURI := str.Split(s, ",")
	if !rxDataURI.MatchString(dataURI[0]) {
		return false
	}
	return IsBase64(dataURI[1])
}

// IsMagnetURI checks if a string is valid magnet URI
func IsMagnetURI(s string) bool {
	return rxMagnetURI.MatchString(s)
}

// IsDNSName will validate the given string as a DNS name
func IsDNSName(s string) bool {
	if s == "" || len(str.Replace(s, ".", "", -1)) > 255 {
		// constraints already violated
		return false
	}
	return !IsIP(s) && rxDNSName.MatchString(s)
}

// IsHash checks if a string is a hash of type algorithm.
// Algorithm is one of ['md4', 'md5', 'sha1', 'sha256', 'sha384', 'sha512', 'ripemd128', 'ripemd160', 'tiger128', 'tiger160', 'tiger192', 'crc32', 'crc32b']
func IsHash(s string, algorithm string) bool {
	n := 0

	algo := str.ToLower(algorithm)
	switch algo {
	case "crc32", "crc32b":
		n = 8
	case "md5", "md4", "ripemd128", "tiger128":
		n = 32
	case "sha1", "ripemd160", "tiger160":
		n = 40
	case "tiger192":
		n = 48
	case "sha256":
		n = 64
	case "sha384":
		n = 96
	case "sha512":
		n = 128
	default:
		return false
	}

	if len(s) != n {
		return false
	}
	return IsAlphaNumeric(s)
}

// IsSHA512 checks is a string is a SHA512 hash. Alias for `IsHash(s, "sha512")`
func IsSHA512(s string) bool {
	return IsHash(s, "sha512")
}

// IsSHA384 checks is a string is a SHA384 hash. Alias for `IsHash(s, "sha384")`
func IsSHA384(s string) bool {
	return IsHash(s, "sha384")
}

// IsSHA256 checks is a string is a SHA256 hash. Alias for `IsHash(s, "sha256")`
func IsSHA256(s string) bool {
	return IsHash(s, "sha256")
}

// IsTiger192 checks is a string is a Tiger192 hash. Alias for `IsHash(s, "tiger192")`
func IsTiger192(s string) bool {
	return IsHash(s, "tiger192")
}

// IsTiger160 checks is a string is a Tiger160 hash. Alias for `IsHash(s, "tiger160")`
func IsTiger160(s string) bool {
	return IsHash(s, "tiger160")
}

// IsRipeMD160 checks is a string is a RipeMD160 hash. Alias for `IsHash(s, "ripemd160")`
func IsRipeMD160(s string) bool {
	return IsHash(s, "ripemd160")
}

// IsSHA1 checks is a string is a SHA-1 hash. Alias for `IsHash(s, "sha1")`
func IsSHA1(s string) bool {
	return IsHash(s, "sha1")
}

// IsTiger128 checks is a string is a Tiger128 hash. Alias for `IsHash(s, "tiger128")`
func IsTiger128(s string) bool {
	return IsHash(s, "tiger128")
}

// IsRipeMD128 checks is a string is a RipeMD128 hash. Alias for `IsHash(s, "ripemd128")`
func IsRipeMD128(s string) bool {
	return IsHash(s, "ripemd128")
}

// IsCRC32 checks is a string is a CRC32 hash. Alias for `IsHash(s, "crc32")`
func IsCRC32(s string) bool {
	return IsHash(s, "crc32")
}

// IsCRC32b checks is a string is a CRC32b hash. Alias for `IsHash(s, "crc32b")`
func IsCRC32b(s string) bool {
	return IsHash(s, "crc32b")
}

// IsMD5 checks is a string is a MD5 hash. Alias for `IsHash(s, "md5")`
func IsMD5(s string) bool {
	return IsHash(s, "md5")
}

// IsMD4 checks is a string is a MD4 hash. Alias for `IsHash(s, "md4")`
func IsMD4(s string) bool {
	return IsHash(s, "md4")
}

// IsDialString validates the given string for usage with the various Dial() functions
func IsDialString(s string) bool {
	if h, p, err := net.SplitHostPort(s); err == nil && h != "" && p != "" && (IsDNSName(h) || IsIP(h)) && IsPort(p) {
		return true
	}

	return false
}

// IsIP checks if a string is either IP version 4 or 6. Alias for `net.ParseIP`
func IsIP(s string) bool {
	return net.ParseIP(s) != nil
}

// IsPort checks if a string represents a valid port
func IsPort(s string) bool {
	if i, err := strconv.Atoi(s); err == nil && i > 0 && i < 65536 {
		return true
	}
	return false
}

// IsIPv4 checks if the string is an IP version 4.
func IsIPv4(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && str.Contains(s, ".")
}

// IsIPv6 checks if the string is an IP version 6.
func IsIPv6(s string) bool {
	ip := net.ParseIP(s)
	return ip != nil && str.Contains(s, ":")
}

// IsCIDR checks if the string is an valid CIDR notiation (IPV4 & IPV6)
func IsCIDR(s string) bool {
	_, _, err := net.ParseCIDR(s)
	return err == nil
}

// IsMAC checks if a string is valid MAC address.
// Possible MAC formats:
// 01:23:45:67:89:ab
// 01:23:45:67:89:ab:cd:ef
// 01-23-45-67-89-ab
// 01-23-45-67-89-ab-cd-ef
// 0123.4567.89ab
// 0123.4567.89ab.cdef
func IsMAC(s string) bool {
	_, err := net.ParseMAC(s)
	return err == nil
}

// IsHost checks if the string is a valid IP (both v4 and v6) or a valid DNS name
func IsHost(s string) bool {
	return IsIP(s) || IsDNSName(s)
}

// IsLatitude checks if a string is valid latitude.
func IsLatitude(s string) bool {
	return rxLatitude.MatchString(s)
}

// IsLongitude checks if a string is valid longitude.
func IsLongitude(s string) bool {
	return rxLongitude.MatchString(s)
}

// IsIMEI checks if a string is valid IMEI
func IsIMEI(s string) bool {
	return rxIMEI.MatchString(s)
}

// IsIMSI checks if a string is valid IMSI
func IsIMSI(s string) bool {
	if !rxIMSI.MatchString(s) {
		return false
	}

	mcc, err := strconv.ParseInt(s[0:3], 10, 32)
	if err != nil {
		return false
	}

	switch mcc {
	case 202, 204, 206, 208, 212, 213, 214, 216, 218, 219:
	case 220, 221, 222, 226, 228, 230, 231, 232, 234, 235:
	case 238, 240, 242, 244, 246, 247, 248, 250, 255, 257:
	case 259, 260, 262, 266, 268, 270, 272, 274, 276, 278:
	case 280, 282, 283, 284, 286, 288, 289, 290, 292, 293:
	case 294, 295, 297, 302, 308, 310, 311, 312, 313, 314:
	case 315, 316, 330, 332, 334, 338, 340, 342, 344, 346:
	case 348, 350, 352, 354, 356, 358, 360, 362, 363, 364:
	case 365, 366, 368, 370, 372, 374, 376, 400, 401, 402:
	case 404, 405, 406, 410, 412, 413, 414, 415, 416, 417:
	case 418, 419, 420, 421, 422, 424, 425, 426, 427, 428:
	case 429, 430, 431, 432, 434, 436, 437, 438, 440, 441:
	case 450, 452, 454, 455, 456, 457, 460, 461, 466, 467:
	case 470, 472, 502, 505, 510, 514, 515, 520, 525, 528:
	case 530, 536, 537, 539, 540, 541, 542, 543, 544, 545:
	case 546, 547, 548, 549, 550, 551, 552, 553, 554, 555:
	case 602, 603, 604, 605, 606, 607, 608, 609, 610, 611:
	case 612, 613, 614, 615, 616, 617, 618, 619, 620, 621:
	case 622, 623, 624, 625, 626, 627, 628, 629, 630, 631:
	case 632, 633, 634, 635, 636, 637, 638, 639, 640, 641:
	case 642, 643, 645, 646, 647, 648, 649, 650, 651, 652:
	case 653, 654, 655, 657, 658, 659, 702, 704, 706, 708:
	case 710, 712, 714, 716, 722, 724, 730, 732, 734, 736:
	case 738, 740, 742, 744, 746, 748, 750, 995:
		return true
	default:
		return false
	}
	return true
}
