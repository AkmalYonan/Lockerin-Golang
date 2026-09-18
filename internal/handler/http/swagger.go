package http

import (
	"fmt"
	"net/http"

	"lockerin-backend/docs"
)

// SwaggerHandler serves Swagger UI and the OpenAPI specification.
type SwaggerHandler struct{}

// NewSwaggerHandler creates a new instance of SwaggerHandler.
func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

// ServeOpenAPISpec serves the raw OpenAPI YAML file.
func (h *SwaggerHandler) ServeOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(docs.OpenAPISpec)
}

// ServeUI renders the Swagger UI HTML page.
func (h *SwaggerHandler) ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, swaggerHTML)
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Lockerin API - Swagger Documentation</title>
  <link rel="stylesheet" type="text/css" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css" />
  <link rel="icon" type="image/png" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/favicon-32x32.png" sizes="32x32" />
  <style>
    html {
      box-sizing: border-box;
      overflow-y: scroll;
    }
    *, *:before, *:after {
      box-sizing: inherit;
    }
    body {
      margin: 0;
      background: #f8fafc;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    }
    .custom-header {
      background: #0f172a;
      color: #ffffff;
      padding: 14px 24px;
      display: flex;
      align-items: center;
      justify-content: space-between;
      border-bottom: 3px solid #2563eb;
      box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
    }
    .custom-header .brand {
      display: flex;
      align-items: center;
      gap: 12px;
      font-weight: 700;
      font-size: 1.15rem;
      letter-spacing: -0.025em;
    }
    .custom-header .badge {
      background: #2563eb;
      color: #fff;
      font-size: 0.75rem;
      padding: 3px 8px;
      border-radius: 9999px;
      font-weight: 600;
    }
    .custom-header .links a {
      color: #94a3b8;
      text-decoration: none;
      font-size: 0.875rem;
      margin-left: 16px;
      font-weight: 500;
      transition: color 0.2s;
    }
    .custom-header .links a:hover {
      color: #38bdf8;
    }
    .swagger-ui .topbar {
      display: none;
    }
    .swagger-ui .information-container {
      margin-top: 15px;
    }
    .swagger-ui .scheme-container {
      background: #ffffff;
      box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.05);
      border-radius: 8px;
      margin: 20px 0;
      padding: 15px 20px;
    }
  </style>
</head>
<body>
  <header class="custom-header">
    <div class="brand">
      <span>🔐 Lockerin Modular Backend API</span>
      <span class="badge">OpenAPI 3.0</span>
    </div>
    <div class="links">
      <a href="/docs/openapi.yaml" target="_blank" download="openapi.yaml">📥 Download Spec (YAML)</a>
      <a href="/health" target="_blank">🩺 Health Check</a>
    </div>
  </header>

  <div id="swagger-ui"></div>

  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js" charset="UTF-8"></script>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-standalone-preset.js" charset="UTF-8"></script>
  <script>
    window.onload = function() {
      // Dynamically resolve the active server base URL from the current browser origin.
      // This ensures "Try it out" works correctly in any environment:
      // - http://localhost:8080 during local development
      // - https://lockerin-golang.vercel.app on Vercel
      // - any custom domain in the future
      var activeBaseURL = window.location.origin + "/api/v1";

      window.ui = SwaggerUIBundle({
        url: "/docs/openapi.yaml",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout",
        persistAuthorization: true,
        displayRequestDuration: true,
        filter: true,
        tryItOutEnabled: true,
        docExpansion: "list",
        defaultModelsExpandDepth: 1,
        defaultModelExpandDepth: 1,
        // Override server URL to always use the current domain
        requestInterceptor: function(request) {
          // Replace any hardcoded server URL (localhost or other) with the active domain
          if (request.url) {
            request.url = request.url.replace(
              /^https?:\/\/[^\/]+\/api\/v1/,
              activeBaseURL
            );
          }
          return request;
        },
        // Inject current domain as the first/default server option
        onComplete: function() {
          // Update the server selector dropdown to show the active URL
          var serverSelect = document.querySelector('.servers select');
          if (serverSelect) {
            // Add active domain as first option if not already present
            var alreadyPresent = false;
            for (var i = 0; i < serverSelect.options.length; i++) {
              if (serverSelect.options[i].value === activeBaseURL) {
                alreadyPresent = true;
                serverSelect.selectedIndex = i;
                break;
              }
            }
            if (!alreadyPresent) {
              var opt = document.createElement('option');
              opt.value = activeBaseURL;
              opt.text = activeBaseURL + " (Current)";
              serverSelect.insertBefore(opt, serverSelect.firstChild);
              serverSelect.selectedIndex = 0;
            }
          }
        }
      });
    };
  </script>
</body>
</html>
`

