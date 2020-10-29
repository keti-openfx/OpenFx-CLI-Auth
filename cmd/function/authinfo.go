package function

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/keti-openfx/openfx-cli/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type Access_info struct {
	Client_id  string `json:"client_id"`
	Expires_in int    `json:"expires_in"`
	Scope      string `json:"scope"`
	User_id    string `json:"user_id"`
	Grade      string `json:"grade"`
}

func init() {
	_, _, access_token, err := config.LookupAuthConfig()
	if err != nil {
		log.Println(err)
	}

	authinfoCmd.Flags().StringVar(&token, "token", access_token, "AccessToken")
	authinfoCmd.Flags().StringVar(&authServerURL, "server", config.DefaultOAuth2Server, "Auth Server")

}

var authinfoCmd = &cobra.Command{
	Use:   `authinfo --token <token value> `,
	Short: "Validate the access token.",
	Long: `
	Validate the access token.
	`,
	Example: `openfx-cli function authinfo --token <Access Token> 
    `,
	PreRunE: preRunAuthInfo,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runAuthInfo(); err != nil {
			return err
		}
		return nil
	},
}

func preRunAuthInfo(cmd *cobra.Command, args []string) error {
	return nil
}

// auth server 키고 handler 처리(state 받고 login 에 자동으로 url 이 전송되게끔)도 해야함
func runAuthInfo() error {

	res, err := http.Get(fmt.Sprintf("%s/verify?access_token=%s", authServerURL, token))
	defer res.Body.Close()

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("cannot POST to %s", authURL))
	}
	if res.Body != nil {
		var target Access_info

		tokenData, _ := ioutil.ReadAll(res.Body)

		jsonErr := json.Unmarshal(tokenData, &target)
		if jsonErr != nil {
			// 에러 메시지
			return fmt.Errorf("[Information Error] The client or user information is incorrect. %s")
		}
		fmt.Printf("Cleint ID             :  %v\n", target.Client_id)
		fmt.Printf("User ID               :  %v\n", target.User_id)
		fmt.Printf("Allowed resources     :  %v\n", target.Scope)
		fmt.Printf("Token valid time(sec) :  %v\n", target.Expires_in)
		fmt.Printf("Grade                 :  %v\n", target.Grade)
	}

	return nil
}
