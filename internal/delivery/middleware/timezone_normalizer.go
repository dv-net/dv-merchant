package middleware

import (
	"bytes"
	"time"

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v3"
)

// deprecatedTimezones maps deprecated IANA timezone names to their canonical equivalents.
// Some OS distributions (e.g., Ubuntu 24.04) remove deprecated aliases from tzdata.
var deprecatedTimezones = map[string]string{
	"Asia/Calcutta":        "Asia/Kolkata",
	"Asia/Saigon":          "Asia/Ho_Chi_Minh",
	"Asia/Katmandu":        "Asia/Kathmandu",
	"Asia/Rangoon":         "Asia/Yangon",
	"Asia/Dacca":           "Asia/Dhaka",
	"Asia/Thimbu":          "Asia/Thimphu",
	"Asia/Ujung_Pandang":   "Asia/Makassar",
	"Asia/Ulan_Bator":      "Asia/Ulaanbaatar",
	"Asia/Chungking":       "Asia/Chongqing",
	"Asia/Macao":           "Asia/Macau",
	"Asia/Tel_Aviv":        "Asia/Jerusalem",
	"Asia/Ashkhabad":       "Asia/Ashgabat",
	"Europe/Kiev":          "Europe/Kyiv",
	"Atlantic/Faeroe":      "Atlantic/Faroe",
	"Pacific/Ponape":       "Pacific/Pohnpei",
	"Pacific/Truk":         "Pacific/Chuuk",
	"Pacific/Samoa":        "Pacific/Pago_Pago",
	"America/Buenos_Aires": "America/Argentina/Buenos_Aires",
	"America/Indianapolis": "America/Indiana/Indianapolis",
	"America/Louisville":   "America/Kentucky/Louisville",
	"America/Porto_Acre":   "America/Rio_Branco",
	"America/Santa_Isabel": "America/Tijuana",
	"America/Virgin":       "America/St_Thomas",
	"Africa/Asmera":        "Africa/Asmara",
	"Africa/Timbuktu":      "Africa/Bamako",
	"America/Godthab":      "America/Nuuk",
}

// TimezoneNormalizer middleware normalizes deprecated timezone names to their canonical equivalents.
// It modifies the request body before it reaches the handler, replacing deprecated timezone
// values (e.g., "Asia/Calcutta") with their canonical names (e.g., "Asia/Kolkata").
func TimezoneNormalizer() fiber.Handler {
	return func(c fiber.Ctx) error {
		if c.Method() != fiber.MethodPost && c.Method() != fiber.MethodPut && c.Method() != fiber.MethodPatch {
			return c.Next()
		}

		contentType := c.Get(fiber.HeaderContentType)
		if !bytes.HasPrefix([]byte(contentType), []byte(fiber.MIMEApplicationJSON)) {
			return c.Next()
		}

		body := c.Body()
		if len(body) == 0 {
			return c.Next()
		}

		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			return c.Next()
		}

		modified := false
		for _, field := range []string{"location", "timezone"} {
			if val, ok := data[field].(string); ok {
				if canonical, exists := deprecatedTimezones[val]; exists {
					// Only convert if deprecated timezone is not supported by the OS
					if _, err := time.LoadLocation(val); err != nil {
						data[field] = canonical
						modified = true
					}
				}
			}
		}

		if !modified {
			return c.Next()
		}

		newBody, err := json.Marshal(data)
		if err != nil {
			return c.Next()
		}

		c.Request().SetBody(newBody)
		c.Request().Header.SetContentLength(len(newBody))
		return c.Next()
	}
}
