package templater

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"reflect"
	"regexp"
	"strings"

	"github.com/dv-net/dv-merchant/internal/service/setting"
	"github.com/dv-net/dv-merchant/internal/settings"
	"github.com/dv-net/dv-merchant/internal/util"
	"github.com/dv-net/dv-merchant/pkg/logger"

	"github.com/cbroglie/mustache"
	assets "github.com/dv-net/email-template"
	"github.com/jellydator/ttlcache/v3"
	"github.com/samber/lo"
	"golang.org/x/text/language"
)

type ITemplaterService interface {
	AssembleTemplate(string) (*mustache.Template, error)
	AssembleEmail(IEmailPayload) (*bytes.Buffer, error)
	ClearCache()
	ClearCacheForTemplate(string)
}

var _ ITemplaterService = (*Service)(nil)

type Service struct {
	logger            logger.Logger
	provider          mustache.PartialProvider
	localizationFiles map[string][]byte
	templates         map[string]*mustache.Template
	templateCache     *ttlcache.Cache[string, *mustache.Template]
	settingSvc        setting.ISettingService
	mailerSettings    *settings.MailerSettings
}

func New(ctx context.Context, logger logger.Logger, settings setting.ISettingService) ITemplaterService {
	svc := &Service{
		logger:        logger,
		settingSvc:    settings,
		templateCache: ttlcache.New[string, *mustache.Template](),
	}

	if err := svc.initializeSettings(ctx); err != nil {
		logger.Error("failed to initialize settings", err)
		return nil
	}

	if err := svc.parsePartialsFS(); err != nil {
		logger.Error("failed to initialize provider", err)
	}

	if err := svc.initializeLocalization(); err != nil {
		logger.Error("failed to initialize localization", err)
	}

	if err := svc.preloadTemplates(); err != nil {
		logger.Error("failed to preload templates", err)
	}

	// Start cache cleanup goroutine
	go svc.templateCache.Start()

	svc.logger.Info(
		"Templater service initialized",
		"templates_count",
		len(svc.templates),
		"localization_files_count", len(svc.localizationFiles),
		"template_cache_initialized", true,
	)
	return svc
}

// EmailHeader represents email header information
type EmailHeader struct {
	Sender   string
	Receiver string
	Subject  string
}

// Build creates an email with proper headers
func (h *EmailHeader) Build(body *bytes.Buffer) (*bytes.Buffer, error) {
	headerBuffer := new(bytes.Buffer)
	headerBuffer.WriteString("MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n")
	fmt.Fprintf(headerBuffer, "From: %s\n", h.Sender)
	fmt.Fprintf(headerBuffer, "To: %s\n", h.Receiver)
	fmt.Fprintf(headerBuffer, "Subject: %s\n\n", h.Subject)
	_, err := headerBuffer.ReadFrom(body)
	return headerBuffer, err
}

// AssembleTemplate assembles a template from partials, leaving placeholders for the actual data
//
// Returned mustache.Template can be rendered multiple times with new data to produce the final email
// AssembleTemplate assembles a template from partials, leaving placeholders for the actual data
//
// Returned mustache.Template can be rendered multiple times with new data to produce the final email
func (o *Service) AssembleTemplate(templateName string) (*mustache.Template, error) {
	rootPartial, err := o.provider.Get(templateName)
	if rootPartial == "" && err == nil {
		return nil, fmt.Errorf("%s %w", templateName, ErrPartialNotFound)
	}
	if err != nil {
		return nil, err
	}

	// Assemble template from partials via string since ParseFile uses os.ReadFile
	template, err := mustache.ParseStringPartials(rootPartial, o.provider)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	return template, nil
}

// generateCacheKey creates a cache key for language-specific templates
func (o *Service) generateCacheKey(templateName, language string) string {
	normalizedLang := util.ParseLanguageTag(language).String()
	return fmt.Sprintf("%s:%s", templateName, normalizedLang)
}

// snakeToCamelCase converts snake_case string to camelCase
func snakeToCamelCase(snake string) string {
	// Use regex to find underscores followed by letters
	re := regexp.MustCompile(`_([a-z])`)
	camel := re.ReplaceAllStringFunc(snake, func(match string) string {
		// Remove underscore and capitalize the letter
		return strings.ToUpper(match[1:])
	})
	return camel
}

// transformMapKeysToCamelCase recursively transforms all keys in a map from snake_case to camelCase
func transformMapKeysToCamelCase(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	for key, value := range data {
		camelKey := snakeToCamelCase(key)

		// Recursively transform nested maps
		switch v := value.(type) {
		case map[string]interface{}:
			result[camelKey] = transformMapKeysToCamelCase(v)
		case []interface{}:
			// Handle arrays of maps
			var transformedArray []interface{}
			for _, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					transformedArray = append(transformedArray, transformMapKeysToCamelCase(itemMap))
				} else {
					transformedArray = append(transformedArray, item)
				}
			}
			result[camelKey] = transformedArray
		default:
			result[camelKey] = value
		}
	}

	return result
}

// getLocalizedTemplate retrieves a cached template or creates and caches it per language
func (o *Service) getLocalizedTemplate(templateName, language string) (*mustache.Template, error) {
	cacheKey := o.generateCacheKey(templateName, language)

	// Check cache first
	if cachedTemplate := o.templateCache.Get(cacheKey); cachedTemplate != nil {
		return cachedTemplate.Value(), nil
	}

	// Template not in cache, assemble it
	template, err := o.AssembleTemplate(templateName)
	if err != nil {
		return nil, wrapError(err, "failed to assemble template %s", templateName)
	}

	// Cache the template with language-specific key
	// The actual localization is applied during rendering with the real payload
	o.templateCache.Set(cacheKey, template, ttlcache.NoTTL)

	return template, nil
}

// ClearCache clears the template cache for all languages
func (o *Service) ClearCache() {
	o.templateCache.DeleteAll()
	o.logger.Info("Template cache cleared")
}

// ClearCacheForTemplate clears cache for a specific template across all languages
func (o *Service) ClearCacheForTemplate(templateName string) {
	items := o.templateCache.Items()
	for key := range items {
		if strings.HasPrefix(key, templateName+":") {
			o.templateCache.Delete(key)
		}
	}
	o.logger.Info("Template cache cleared for template", "template", templateName)
}

// applyLocalization applies localization data to the payload based on language
func (o *Service) applyLocalization(payload IEmailPayload) error {
	locale := o.localizationFiles[util.ParseLanguageTag(payload.GetLanguage()).String()]

	// If no localization file is found, fallback to English
	if locale == nil {
		locale = o.localizationFiles[language.English.String()]
	}

	// Apply localization directly to payload using existing snake_case JSON tags
	// The camelCase transformation will happen later in payloadToMap for mustache
	if err := json.Unmarshal(locale, payload); err != nil {
		return wrapError(err, "failed to unmarshal localization file into payload")
	}

	return nil
}

// payloadToMap converts a payload struct to a map with camelCase keys for mustache templates
func (o *Service) payloadToMap(payload IEmailPayload) (map[string]interface{}, error) {
	// First marshal the payload to JSON (this uses the existing JSON tags)
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, wrapError(err, "failed to marshal payload to JSON")
	}

	// Unmarshal into a map
	var payloadMap map[string]interface{}
	if err := json.Unmarshal(payloadJSON, &payloadMap); err != nil {
		return nil, wrapError(err, "failed to unmarshal payload JSON to map")
	}

	// Add method results to map using reflection
	if err := addMethodsToMap(payload, payloadMap); err != nil {
		return nil, wrapError(err, "failed to add methods to map")
	}

	// Transform keys to camelCase for mustache
	camelMap := transformMapKeysToCamelCase(payloadMap)

	return camelMap, nil
}

// addMethodsToMap uses reflection to find string and boolean methods and add their results to the map
func addMethodsToMap(payload interface{}, targetMap map[string]interface{}) error { //nolint:unparam
	v := reflect.ValueOf(payload)
	t := reflect.TypeOf(payload)

	// Handle pointer types
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// Iterate through all methods
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)
		methodValue := v.Method(i)

		// Only include exported methods that:
		// - Take no parameters (besides receiver)
		// - Return exactly one value of type string or bool
		if method.Type.NumIn() == 1 && // Only receiver parameter
			method.Type.NumOut() == 1 && // Exactly one return value
			(method.Type.Out(0).Kind() == reflect.String || method.Type.Out(0).Kind() == reflect.Bool) {
			// Call the method
			results := methodValue.Call(nil)
			if len(results) == 1 {
				// Convert method name to snake_case for map key
				methodName := camelToSnakeCase(method.Name)
				targetMap[methodName] = results[0].Interface()
			}
		}
	}

	return nil
}

// camelToSnakeCase converts CamelCase to snake_case
func camelToSnakeCase(str string) string {
	var result strings.Builder
	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// renderTemplate renders the template with payload data, handling nested localization
func (o *Service) renderTemplate(template *mustache.Template, payload IEmailPayload) (*bytes.Buffer, error) {
	// Convert payload to map with camelCase keys for mustache
	payloadMap, err := o.payloadToMap(payload)
	if err != nil {
		return nil, wrapError(err, "failed to convert payload to map")
	}

	buffer := new(bytes.Buffer)
	if err := template.FRender(buffer, payloadMap); err != nil {
		return nil, wrapError(err, "failed to render template")
	}

	renderedEmail, err := mustache.ParseStringRaw(buffer.String(), true)
	if err != nil {
		return nil, wrapError(err, "failed to parse email template")
	}

	// Re-render email template to add nested localization
	// TODO: Find a better way to handle nested localization, this is a workaround
	buffer.Reset()
	if err := renderedEmail.FRender(buffer, payloadMap); err != nil {
		return nil, wrapError(err, "failed to re-render email template")
	}

	return buffer, nil
}

// AssembleEmail creates a complete email from template and payload data
func (o *Service) AssembleEmail(payload IEmailPayload) (*bytes.Buffer, error) {
	if payload == nil {
		return nil, ErrPayloadNil
	}

	template, err := o.getLocalizedTemplate(payload.GetName(), payload.GetLanguage())
	if err != nil {
		return nil, wrapError(err, "failed to get localized template %s:%s", payload.GetName(), payload.GetLanguage())
	}

	if err := o.applyLocalization(payload); err != nil {
		return nil, wrapError(err, "failed to apply localization")
	}

	buffer, err := o.renderTemplate(template, payload)
	if err != nil {
		return nil, wrapError(err, "failed to render template")
	}

	header := &EmailHeader{
		Sender:   o.mailerSettings.MailerSender,
		Receiver: payload.GetUserEmail(),
		Subject:  payload.GetSubject(),
	}
	buffer, err = header.Build(buffer)
	if err != nil {
		return nil, wrapError(err, "failed to add email header")
	}

	return buffer, nil
}

func (o *Service) initializeSettings(ctx context.Context) error {
	mailerSettings, err := o.settingSvc.GetMailerSettings(ctx)
	if err != nil {
		return wrapError(err, "failed to get mailer settings")
	}
	o.mailerSettings = mailerSettings
	return nil
}

func (o *Service) parsePartialsFS() error {
	entries, err := fs.ReadDir(assets.Assets, "mustache")
	if err != nil {
		return err
	}

	paths := lo.Uniq(lo.FilterMap(entries, func(dir fs.DirEntry, _ int) (string, bool) {
		if dir.IsDir() {
			return path.Join("mustache", dir.Name()), true
		}
		return "", false
	}))

	provider := &EmbeddedProvider{
		FS:    assets.Assets,
		Paths: paths,
	}
	o.provider = provider
	o.logger.Info("Partial provider initialized", "paths", paths)
	return nil
}

func (o *Service) initializeLocalization() error {
	localesDir, err := fs.ReadDir(assets.Assets, "i18n")
	if err != nil {
		return err
	}

	localeFiles := make(map[string][]byte)
	for _, locale := range localesDir {
		if !locale.IsDir() && path.Ext(locale.Name()) == ".json" {
			data, err := assets.Assets.ReadFile("i18n/" + locale.Name())
			if err != nil {
				return fmt.Errorf("failed to read file %s: %w", locale.Name(), err)
			}
			localeFiles[strings.TrimSuffix(locale.Name(), path.Ext(locale.Name()))] = data
		}
	}
	o.localizationFiles = localeFiles
	return nil
}

func (o *Service) preloadTemplates() error {
	if o.templates == nil {
		o.templates = make(map[string]*mustache.Template)
	}
	for k := range validPartials {
		template, err := o.AssembleTemplate(k)
		if err != nil {
			return fmt.Errorf("failed to assemble template %s: %w", k, err)
		}
		o.templates[k] = template
	}
	return nil
}
