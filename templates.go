package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

type tokenTmplData struct {
	IDToken      string
	RefreshToken string
	RedirectURL  string
	Claims       interface{}
	ClientSecret string
	ApiServer    string
	ApiCa        string
}

var tokenTmpl = template.Must(template.New("token.html").Parse(`<html>
  <head>
<style type="text/css">
.form-style-5{
    max-width: 1500px;
    padding: 20px 100px;
    background: #f4f7f8;
    margin: 10px auto;
    background: #f4f7f8;
    border-radius: 8px;
    font-family: Georgia, "Times New Roman", Times, serif;
}
</style>
    <style>
/* make pre wrap */
pre {
 white-space: pre-wrap;       /* css-3 */
 white-space: -moz-pre-wrap;  /* Mozilla, since 1999 */
 white-space: -pre-wrap;      /* Opera 4-6 */
 white-space: -o-pre-wrap;    /* Opera 7 */
 word-wrap: break-word;       /* Internet Explorer 5.5+ */
}
    </style>
    <link rel="stylesheet" href="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/styles/default.min.css">
    <script src="//cdnjs.cloudflare.com/ajax/libs/highlight.js/9.12.0/highlight.min.js"></script>
  </head>
  <body>
  <div class="form-style-5">
  <p>Copy this in your ~/.kube/config file:</p>
  <pre><code class="hljs yaml">
apiVersion: v1
clusters:
- cluster:
    server: {{ .ApiServer }}
    certificate-authority-data: {{ .ApiCa }}
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: {{ .Claims.name }}
  name: {{ .Claims.name }}@kubernetes
current-context: {{ .Claims.name }}@kubernetes
kind: Config
preferences: {}
users:
- name: {{ .Claims.name }}
  user:
    auth-provider:
      config:
        client-id: {{ .Claims.aud }}
        client-secret: {{ .ClientSecret }}
        id-token: {{ .IDToken }}
        idp-issuer-url: {{ .Claims.iss }}
        refresh-token: {{ .RefreshToken }}
      name: oidc
  </code></pre>
  </body>
</html>
`))

func renderToken(w http.ResponseWriter, redirectURL, idToken, refreshToken string, claims []byte, clientSecret string, apiServer string, apiCa string) {
	var json_claims map[string]interface{}
	if err := json.Unmarshal(claims, &json_claims); err != nil {
		panic(err)
	}
	renderTemplate(w, tokenTmpl, tokenTmplData{
		IDToken:      idToken,
		RefreshToken: refreshToken,
		RedirectURL:  redirectURL,
		Claims:       json_claims,
		ClientSecret: clientSecret,
		ApiServer:    apiServer,
		ApiCa:        apiCa,
	})
}

func renderTemplate(w http.ResponseWriter, tmpl *template.Template, data interface{}) {
	err := tmpl.Execute(w, data)
	if err == nil {
		return
	}

	switch err := err.(type) {
	case *template.Error:
		// An ExecError guarantees that Execute has not written to the underlying reader.
		log.Printf("Error rendering template %s: %s", tmpl.Name(), err)

		// TODO(ericchiang): replace with better internal server error.
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	default:
		// An error with the underlying write, such as the connection being
		// dropped. Ignore for now.
	}
}
