package main;

import (
	"net/http"
	"io/ioutil"
	"encoding/json"
)
const UPDATE_URL = "https://krake.one/files/DiscordConsole.json";

type updateObj struct{
	Version string `json:"version"`
	Url string `json:"url"`
	UpdateAvailable bool
}

func checkUpdate() (updateObj, error){
	res, err := http.Get(UPDATE_URL);
	if(err != nil){
		return updateObj{}, err;
	}
	defer res.Body.Close();

	content, err := ioutil.ReadAll(res.Body);
	if(err != nil){
		return updateObj{}, err;
	}

	var update updateObj;
	err = json.Unmarshal(content, &update);
	if(err != nil){
		return updateObj{}, err;
	}

	update.UpdateAvailable = update.Version != VERSION;
	return update, nil;
}
