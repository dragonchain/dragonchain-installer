package dragonnet

import (
	"errors"
	"io/ioutil"
	"net/http"
	"time"
)

var tryLimit = 30

func checkMatchmakingRegistration(pubID string, tries int) error {
	// First check that the chain was able to register correctly (has correct dragon net tokens)
	resp, err := http.Get("https://matchmaking.api.dragonchain.com/registration/" + pubID)
	if err != nil {
		return errors.New("Error communicating with matchmaking:\n" + err.Error())
	}
	if resp.StatusCode != 200 {
		if tries > tryLimit {
			return errors.New("Registration could not be found for your chain. Although your chain may be installed and working locally, dragon net support will not work. Check the logs of the transaction processor for more details")
		}
		time.Sleep(1 * time.Second)
		return checkMatchmakingRegistration(pubID, tries+1)
	}
	resp.Body.Close()
	// Now check that the chain is reachable from the greater internet
	resp, err = http.Get("https://matchmaking.api.dragonchain.com/registration/verify/" + pubID + "?source=installscript")
	if err != nil {
		return errors.New("Error communicating with matchmaking:\n" + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("Error reading matchmaking response body:\n" + err.Error())
	}
	if resp.StatusCode != 200 {
		return errors.New("Although registered, dragon net is reporting that the chain is not reachable (did you port-forward correctly)? Dragon net support will not work. Error:\n" + string(body))
	}
	return nil
}

// CheckDragonNetConfiguration checks if a dragonchain is running and connectable via dragon net
func CheckDragonNetConfiguration(pubID string) error {
	if err := checkMatchmakingRegistration(pubID, 0); err != nil {
		return err
	}
	return nil
}
