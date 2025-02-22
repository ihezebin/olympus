package httpserver

import "strings"

type OpenAPIUITemplateType string

const (
	OpenAPIUITemplateSwagger          OpenAPIUITemplateType = "swagger"
	OpenAPIUITemplateRapiDoc          OpenAPIUITemplateType = "rapidoc"
	OpenAPIUITemplateStoplightElement OpenAPIUITemplateType = "stoplight"
	OpenAPIUITemplateRedoc            OpenAPIUITemplateType = "redoc"
)

var uiTemplates = map[OpenAPIUITemplateType]string{
	OpenAPIUITemplateSwagger:          swagger,
	OpenAPIUITemplateRapiDoc:          rapidoc,
	OpenAPIUITemplateStoplightElement: stoplightElement,
	OpenAPIUITemplateRedoc:            redoc,
}

func OpenAPIHTMLUI(t OpenAPIUITemplateType, title string, spec string) string {
	html := uiTemplates[t]
	html = strings.ReplaceAll(html, "{:title}", title)
	html = strings.ReplaceAll(html, "{:spec}", spec)
	return html
}

const swagger = `
<!DOCTYPE html>
<html charset="UTF-8">
<head>
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8">
    <title>{:title} Document [Swagger UI]</title>
    <link type="text/css" rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui.css">
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@3/swagger-ui-bundle.js"></script>
</head>
</html>
<body>
  <div id="ui"></div>
  <script>
    let spec = {:spec};
    let oauth2RedirectUrl;

    let query = window.location.href.indexOf("?");
    if (query > 0) {
        oauth2RedirectUrl = window.location.href.substring(0, query);
    } else {
        oauth2RedirectUrl = window.location.href;
    }

    if (!oauth2RedirectUrl.endsWith("/")) {
        oauth2RedirectUrl += "/";
    }
    oauth2RedirectUrl += "oauth-receiver.html";
    SwaggerUIBundle({
        dom_id: '#ui',
        spec: spec,
        filter: false,
        oauth2RedirectUrl: oauth2RedirectUrl,
    })
  </script>`

const rapidoc = `
<!DOCTYPE html>
<html charset="UTF-8">
  <head>
    <meta http-equiv="Content-Type" content="text/html;charset=utf-8">
    <meta name="viewport" content="width=device-width, minimum-scale=1, initial-scale=1, user-scalable=yes">
    <title>{:title} Document [RapiDoc]</title>
    <script type="module" src="https://cdn.jsdelivr.net/npm/rapidoc/dist/rapidoc-min.min.js"></script>
  </head>
  <style>
    rapi-doc::part(section-navbar) { /* <<< targets navigation bar */
      background: linear-gradient(90deg, #3d4e70, #2e3746);
    }
  </style>
  <body>
    <rapi-doc id="thedoc" 
    theme="dark" 
    primary-color = "#f54c47"
    bg-color = "#2e3746"
    text-color = "#bacdee"
    default-schema-tab="model" 
    allow-search="false"
    allow-advanced-search="true"
    show-info="true" 
    show-header="true" 
    show-components="true" 
    schema-style="table"
    show-method-in-nav-bar="as-colored-block" 
    allow-try="true"
    allow-authentication="true" 
    regular-font="Open Sans" 
    mono-font="Roboto Mono" 
    font-size="large"
    schema-description-expanded="true">
    </rapi-doc>
    <script>
      document.addEventListener('DOMContentLoaded', (event) => {
        let docEl = document.getElementById("thedoc");
        docEl.loadSpec({:spec});
      })
    </script>
  </body>
</html>`

const stoplightElement = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
    <title>{:title} Document [Elements]</title>
  
    <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
    <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
  </head>
  <body>
    <elements-api id="doc" router="hash" />
  </body>

  <script>
    (async() => {
      let doc = document.getElementById("doc");
      doc.apiDescriptionDocument = {:spec};
    })()
  </script>
</html>`

const redoc = `
<!DOCTYPE html>
<html>
  <head>
    <title>Redoc</title>
    <!-- needed for adaptive design -->
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">

    <style>
      body {
        margin: 0;
        padding: 0;
      }
    </style>
  </head>
  <body>
    <redoc id="doc"></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"> </script>
    <script>
      (async()=>{
        Redoc.init({:spec}, {}, document.getElementById('doc'))
      })()
    </script>
  </body>
</html>`
