
package main

import (
	"golang.org/x/oauth2"
	"time"
	"strconv"
	"net/http"
	"net/url"
	"log"
	"fmt"
	"encoding/json"
)

type TDAccount struct {
	UserID           string `json:"userid"`
	Alias            string `json:"alias"`
	Email            string `json:"email"`
	Pro              int64  `json:"pro"`
	DateFormat       int64  `json:"dateformat"`
	Timezone         int64  `json:"timezone"`
	HideMonths       int64  `json:"hidemonths"`
	HotListPriority  int64  `json:"hotlistpriority"`
	HotListDueDate   int64  `json:"hotlistduedate"`
	HotListStar      int64  `json:"hotliststar"`
	HotListStatus    int64  `json:"hotliststatus"`
	ShowTabNums      int64  `json:"showtabnums"`
	LastEditTask     int64  `json:"lastedit_task"`
	LastDeleteTask   int64  `json:"lastdelete_task"`
	LastEditFolder   int64  `json:"lastedit_folder"`
	LastEditContext  int64  `json:"lastedit_context"`
	LastEditGoal     int64  `json:"lastedit_goal"`
	LastEditLocation int64  `json:"lastedit_location"`
	LastEditNote     int64  `json:"lastedit_note"`
	LastDeleteNote   int64  `json:"lastdelete_note"`
	LastEditList     int64  `json:"lastedit_list"`
	LastEditOutline  int64  `json:"lastedit_outline"`
}

type ToodledoAccount struct {
	OauthToken  oauth2.Token
	AccountInfo TDAccount
	UserID      string
	LastSync    time.Time
	initialized bool
}

var tda ToodledoAccount
var tdx TDAccount

var oauthTokenRefresh bool   = false
var oauthCodeState    string = "foobar"

var client *http.Client

var oauthConf = &oauth2.Config{
	ClientID:     "toodledo13",
	ClientSecret: "api5672652217650",
	RedirectURL:  OauthHttpRedirectUrl + ":" + strconv.Itoa(OauthHttpRedirectPort) + "/oauthcallback",
	Scopes:       []string{"basic", "tasks", "notes", "write"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  ToodledoUrl + "account/authorize.php",
		TokenURL: ToodledoUrl + "account/token.php",
	},
}

func oauthHttpInit() {
	http.HandleFunc("/oauthcallback", oauthHttpHandleAuthCallback)
	http.ListenAndServe(":"+strconv.Itoa(OauthHttpRedirectPort), nil)
}

func oauthHttpHandleAuthCallback(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	state := r.FormValue("state")

	if state != oauthCodeState {
		log.Fatal("Invalid redirect state\n")
	}

	oauthToken, err := oauthConf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}

	tda.OauthToken = *oauthToken
	TDFileAccountWrite()

	fmt.Printf("Authentication successful!\n")      /* to console */
	w.Write([]byte("Authentication successful!\n")) /* to browser */

	/* XXX trigger exit here... */
	fmt.Printf("Complete! Enter <Ctrl-C> to exit...\n")
}

func newOauth() {
	url := oauthConf.AuthCodeURL(oauthCodeState, oauth2.AccessTypeOffline)
	fmt.Printf("Go here to perform the authorization:\n%v\n", url)

	/* blocks in http server and never returns... */
	oauthHttpInit()
}

func TDOauth() {
	if tda.initialized == true {
		if (time.Since(tda.OauthToken.Expiry).Hours() / 24) >= 30 {
			fmt.Printf("NOTICE: Ouath2 token is too old, must re-authorize.\n")
			newOauth() /* doesn't return */
		}

		if time.Now().After(tda.OauthToken.Expiry) {
			oauthTokenRefresh = true
		}

		client = oauthConf.Client(oauth2.NoContext, &tda.OauthToken)
	} else {
		newOauth() /* doesn't return */
	}
}

func TDGetData(location string, values url.Values, data interface{}) {
	//info, err := client.Get(ToodledoUrl + location)
	info, err := client.PostForm(ToodledoUrl+location, values)
	if err != nil {
		log.Fatal(err)
	}

	defer info.Body.Close()

	decoder := json.NewDecoder(info.Body)
	decoder.Decode(data)

	if oauthTokenRefresh {
		new_token, err := client.Transport.(*oauth2.Transport).Source.Token()
		if err != nil {
			log.Fatal(err)
		}

		tda.OauthToken = *new_token
		TDFileAccountWrite()
		oauthTokenRefresh = false

		/* XXX
		 * Problem if user leaves app open for a very long time and
		 * the token expires. We don't (yet) have a way of noticing
		 * this and caching the new token.
		 */
	}
}

func TDFileAccountExists() bool {
	return FileExists(AccountFileName)
}

func TDFileAccountWrite() {
	FileWrite(AccountFileName, tda)
}

func TDFileAccountRead() {
	FileRead(AccountFileName, &tda)
}

func TDAccountInit() {
	tda.initialized = false
	if TDFileAccountExists() {
		TDFileAccountRead()
		tda.initialized = true
	}
	//DumpJSON(tda)
}

