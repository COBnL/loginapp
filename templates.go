package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
)

var indexTmpl = template.Must(template.New("index.html").Parse(`<html>
  <head>
    <style type="text/css">.form-style-7{max-width:400px;margin:50px auto;background:#fff;border-radius:2px;padding:20px;font-family:Georgia,"Times New Roman",Times,serif}.form-style-7 h1{display:block;text-align:center;padding:0;margin:0 0 20px;color:#5C5C5C;font-size:x-large}.form-style-7 ul{list-style:none;padding:0;margin:0}.form-style-7 li{display:block;padding:9px;border:1px solid #DDD;margin-bottom:30px;border-radius:3px}.form-style-7 li:last-child{border:none;margin-bottom:0;text-align:center}.form-style-7 li>label{display:block;float:left;margin-top:-19px;background:#FFF;height:14px;padding:2px 5px;color:#B9B9B9;font-size:14px;overflow:hidden;font-family:Arial,Helvetica,sans-serif}.form-style-7 input[type=password],.form-style-7 input[type=text],.form-style-7 input[type=date],.form-style-7 input[type=datetime],.form-style-7 input[type=email],.form-style-7 input[type=number],.form-style-7 input[type=search],.form-style-7 input[type=time],.form-style-7 input[type=url],.form-style-7 select,.form-style-7 textarea{box-sizing:border-box;-webkit-box-sizing:border-box;-moz-box-sizing:border-box;width:100%;display:block;outline:0;border:none;height:25px;line-height:25px;font-size:16px;padding:0;font-family:Georgia,"Times New Roman",Times,serif}.form-style-7 li>span{background:#F3F3F3;display:block;padding:3px;margin:0 -9px -9px;text-align:center;color:silver;font-family:Arial,Helvetica,sans-serif;font-size:11px}.form-style-7 textarea{resize:none}.form-style-7 input[type=submit],.form-style-7 input[type=button]{background:#2471FF;border:none;padding:10px 20px;border-bottom:3px solid #5994FF;border-radius:3px;color:#D2E2FF}.form-style-7 input[type=submit]:hover,.form-style-7 input[type=button]:hover{background:#6B9FFF;color:#fff}
</style>
    <script type="text/javascript">function adjust_textarea(t){t.style.height="20px",t.style.height=t.scrollHeight+"px"}</script>
  </head>
  <body>
    <form class="form-style-7" action="/login" method="post">
      <h2>{{ .AppName }}</h2>
      <ul>
        <li>
          <label for="cross_client">Authentication for clients :</label>
          <input type="text" name="cross_client"{{ if .DisableChoices }}readonly="readonly"{{ end }} />
          <span>A comma separated list</span>
        </li>
        <li>
          <label for="extra_scopes">Extra scopes :</label>
          <input type="text" name="extra_scopes" value="{{ .Scopes }}"{{ if .DisableChoices }}readonly="readonly"{{ end }}>
          <span>A comma separated list</span>
        </li>
        <li>
          <label for="offline_access">Offline Access :</label>
          <input type="checkbox" name="offline_access" value="yes" checked>
        </li>
        <li>
          <label for="doc_url">Documentation</label>
          <a href="https://github.com/coreos/dex/blob/master/Documentation/custom-scopes-claims-clients.md">Help</a>
        </li>
        <li>
          <input type="submit" value="Request Token" />
        </li>
      </ul>
    </form>
  </body>
</html>`))

func renderIndex(w http.ResponseWriter, initScopes string, disableChoices bool, appName string) {
	renderTemplate(w, indexTmpl, indexTmplData{
		Scopes:         initScopes,
		AppName:        appName,
		DisableChoices: disableChoices,
	})
}

type indexTmplData struct {
	Scopes         string
	AppName        string
	DisableChoices bool
}

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
