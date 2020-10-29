package function

import (
	"bytes"
	"io/ioutil"

	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/keti-openfx/openfx-cli/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	scope        string
	authURL      string
	clientID     string
	clientSecret string
	userID       string
	userPasswd   string
	grant        string
)

type ClientCredentialsReq struct {
	client_id     string `json:"client_id"`
	client_secret string `json:"client_secret"`
	grant_type    string `json:"grant_type"`
}

type token_data struct {
	Access_token string `json:"access_token"`
	Token_type   string `json:"token_type"`
	Expires_in   int    `json:"expires_in"`
	Scope        string `json:"scope"`
}

// * * * *
// config에서 client-id, secret 불러오기
// config 에는 client-id, secret, token 값이 저장되어야함

func init() {
	cli_id, cli_secret, _, err := config.LookupAuthConfig()
	if err != nil {
		log.Println(err)
	}
	authCmd.Flags().StringVar(&authURL, "auth-url", config.DefaultOAuth2Server, "OAuth2 Authorize URL i.e. Openfx-oauth2 git")
	authCmd.Flags().StringVar(&clientID, "client-id", cli_id, "OAuth2 client_id")
	authCmd.Flags().StringVar(&clientSecret, "client-secret", cli_secret, "OAuth2 client_secret, for use with client_credentials grant")
	authCmd.Flags().StringVar(&userID, "id", "", "Openfx user ID")
	authCmd.Flags().StringVar(&userPasswd, "pwd", "", "Openfx User Password")
	authCmd.Flags().StringVar(&scope, "scope", "user-fn1", "scope for OAuth2 flow - i.e. \"openfx-fn\"")
	authCmd.Flags().StringVar(&grant, "grant", "client_credentials", "grant for OAuth2 flow - either implicit, implicit-id or client_credentials")
}

var authCmd = &cobra.Command{
	Use:   `login --auth-url `,
	Short: "Get a Accesstoken for your OpenFx Oauth2 Server",
	Long: `
	Get a Accesstoken for Authentication procedure to access Openfx
	`,
	Example: `openfx-cli function login --id <user-id> --pwd <user-password> 
    `,
	PreRunE: preRunAuth,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runAuth(); err != nil {
			return err
		}
		return nil
	},
}

func preRunAuth(cmd *cobra.Command, args []string) error {
	return checkValues(authURL, clientID, clientSecret)
}

func checkValues(authURL, clientID, clientSecret string) error {
	if len(authURL) == 0 {
		return fmt.Errorf("--auth-url is required and must be a valid Openfx OAuth2 Server")
	}
	u, uErr := url.Parse(authURL)
	if uErr != nil {
		return fmt.Errorf("--auth-url is an invalid URL: %s", uErr.Error())
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("--auth-url is an invalid URL: %s", u.String())
	}

	if len(clientID) == 0 {
		return fmt.Errorf("--client-id is required")
	}

	if len(clientSecret) == 0 {
		return fmt.Errorf("--clientSecret is required")
	}

	if len(userID) == 0 {
		return fmt.Errorf("--userID is required")
	}

	if len(userPasswd) == 0 {
		return fmt.Errorf("--userPasswd is required")
	}
	return nil
}

// auth server 키고 handler 처리(state 받고 login 에 자동으로 url 이 전송되게끔)도 해야함
func runAuth() error {

	body := ClientCredentialsReq{
		client_id:     clientID,
		client_secret: clientSecret,
		grant_type:    grant,
	}

	bodyBytes, marshalErr := json.Marshal(body)
	if marshalErr != nil {
		return errors.Wrapf(marshalErr, "unable to unmarshal %s", string(bodyBytes))
	}

	buf := bytes.NewBuffer(bodyBytes)
	// url parsing 필요

	relativeUrl := "/token"
	u, err := url.Parse(relativeUrl)
	if err != nil {
		log.Fatal(err)
	}

	queryString := u.Query()

	queryString.Set("grant_type", "client_credentials")
	queryString.Set("client_id", clientID)
	queryString.Set("client_secret", clientSecret)
	queryString.Set("username", userID)
	queryString.Set("password", userPasswd)

	u.RawQuery = queryString.Encode()

	base, err := url.Parse(authURL)
	if err != nil {
		log.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, base.ResolveReference(u).String(), buf)

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot POST to %s", authURL))
	}
	if res.Body != nil {
		defer res.Body.Close()

		tokenData, _ := ioutil.ReadAll(res.Body)

		if res.StatusCode != http.StatusOK {
			// 에러 메세지 업데이트 필요
			return fmt.Errorf("[Information Error] The client information is incorrect.  %s")
		}
		token := token_data{}
		tokenErr := json.Unmarshal(tokenData, &token)
		if tokenErr != nil {
			return errors.Wrapf(tokenErr, "unable to unmarshal token: %s", string(tokenData))
		}
		config.UpdateAuthConfig(clientID, clientSecret, token.Access_token)

		log.Println("successfully completed the certification.\n")
	}

	return nil
}

/* func authAuthcode() error {
	// oauth2 : auth code grant type
	// What you need is interactive client type example ex.web
	var (
		config = oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  "http://10.0.0.91:9090/oauth2",
			Scopes:       []string{scope},
			Endpoint: oauth2.Endpoint{
				AuthURL:   authURL + "/authorize",
				TokenURL:  authURL + "/token",
				AuthStyle: 0,
			},
		}
	)

	url := config.AuthCodeURL("xyz")
	fmt.Printf("Visit this URL in your browser:\n\n%s\n\n", url)

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	//default
	http.HandleFunc("/oauth2", func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()

		if s := r.URL.Query().Get("state"); s != "xyz" {
			http.Error(w, fmt.Sprintf("Invalid state: %s", s), http.StatusUnauthorized)
			return
		}

		code := r.URL.Query().Get("code")

		log.Printf("code : %+v", code)
		log.Printf("config : %+v", config)

		token, err := config.Exchange(ctx, code)
		if err != nil {
			http.Error(w, fmt.Sprintf("Exchange error: %s", err), http.StatusServiceUnavailable)
			return
		}

		tokenJSON, err := json.MarshalIndent(token, "", "  ")
		if err != nil {
			http.Error(w, fmt.Sprintf("Token parse error: %s", err), http.StatusServiceUnavailable)
			return
		}

		// web Page
		w.Write(tokenJSON)

		log.Printf("token : %+v", token)

	})

	server := http.Server{
		Addr: fmt.Sprintf(":%d", listenPort),
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	wg.Wait()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
	return nil
} */
