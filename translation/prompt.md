After installing the module, make this changes:
1. Find TRANSLATION_LOCALES in the .env file and add your supported locales. For example en-US
2. Run `translation-extract` to extract the locales/messages.pot file.
3. Create a folder `locales/en-US` and place a copy of locales/messages.pot as locales/en-US/messages.po.
4. Translate the messages.po file.
5. Run `msgfmt -o locales/en-US/messages.mo locales/en-US/messages.po` to compile the locales/en-US/messages.po file into locales/en-US/messages.mo.
6. Add the following code to the main.go file:
```go
package main
import (
	"github.com/go-modulus/modulus/translation"
)
func main() {
	// ...
    http.OverrideMiddlewarePipeline(
        http.NewModule(),
        func(
            ...
            translationMd *translation.Middleware,
        ) *http.Pipeline {
            return &http.Pipeline{
                Middlewares: []http.Middleware{
                    ...,
                    translationMd.Middleware,
                },
            }
        },

```
To make all error responses localized.