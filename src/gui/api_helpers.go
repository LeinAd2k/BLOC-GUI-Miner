package gui

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

// GetPoolList returns the list of pools available to the GUI miner
func (gui *GUI) GetPoolList() ([]PoolData, error) {
	// gui.logger.Info("GetPoolList")
	// gui.logger.Info(fmt.Sprintf("%s/pool-list?allowed=true&coin=%s", gui.config.APIEndpoint, gui.config.CoinType))

	var pools []PoolData
	resp, err := http.Get(fmt.Sprintf("%s/pool-list?allowed=true&coin=%s", gui.config.APIEndpoint, gui.config.CoinType))
	if err != nil {
		return pools, err
	}
	err = json.NewDecoder(resp.Body).Decode(&pools)
	if err != nil {
		return pools, err
	}
	return pools, nil
}

// GetPool returns a single pool's information
func (gui *GUI) GetPool(id int) (PoolData, error) {
	// gui.logger.Info("GetPool")
	// gui.logger.Info(fmt.Sprintf("%s/pool/%d?coin=%s", gui.config.APIEndpoint, id, gui.config.CoinType))

	var pool PoolData
	resp, err := http.Get(fmt.Sprintf("%s/pool/%d?coin=%s", gui.config.APIEndpoint, id, gui.config.CoinType))
	if err != nil {
		return pool, err
	}
	err = json.NewDecoder(resp.Body).Decode(&pool)
	if err != nil {
		return pool, err
	}
	return pool, nil
}

// SaveConfig saves the configuration to disk
func (gui *GUI) SaveConfig(config Config) error {
	configBytes, err := json.MarshalIndent(&config, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(
		filepath.Join(gui.workingDir, "config.json"),
		configBytes,
		0644)
	if err != nil {
		return err
	}
	return nil
}

// GetCoinContent returns the content for all coins
func (gui *GUI) GetCoinContentJson() (string, error) {
	t := time.Now()
	resp, err := http.Get(fmt.Sprintf("https://raw.githubusercontent.com/furiousteam/BLOC-GUI-Miner/master/coins/content.json?_=%d", t.Unix()))
	// resp, err := http.Get(fmt.Sprintf("https://bloc.money/miner-content.json?_=%d", t.Unix()))
	if err != nil {
		return "", err
	}
	statBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// gui.logger.Info(fmt.Sprintf("%s", statBytes))

	var content coinsContentJson
	err = json.Unmarshal(statBytes, &content)
	if err != nil {
		return "", err
	}
	statBytes, err = json.Marshal(&content)
	if err != nil {
		return "", err
	}
	// gui.logger.Info(fmt.Sprintf("%s", statBytes))
	return string(statBytes), nil
}

// GetStats returns stats for the interface. It requires the miner's
// hashrate to calculate BLOC per dat
func (gui *GUI) GetStats(

	poolID int,
	hashrate float64,
	mid string) (string, error) {
	// gui.logger.Info("GetStats")
	// gui.logger.Info(fmt.Sprintf("%s/stats?pool=%d&hr=%.2f&mid=%s&coin=%s", gui.config.APIEndpoint, poolID, hashrate, mid, gui.config.CoinType))

	if mid == "" || poolID == 0 {
		return "", errors.New("No data yet")
	}
	resp, err := http.Get(
		fmt.Sprintf("%s/stats?pool=%d&hr=%.2f&mid=%s&coin=%s",
			gui.config.APIEndpoint,
			poolID,
			hashrate,
			mid,
			gui.config.CoinType))
	if err != nil {
		return "", err
	}
	statBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var stats GlobalStats
	err = json.Unmarshal(statBytes, &stats)
	if err != nil {
		return "", err
	}

	poolTemplate, err := gui.GetPoolTemplate()
	if err != nil {
		log.Fatalf("Unable to load pool template: '%s'", err)
	}
	poolData := PoolData{
		ID:        stats.Pool.ID,
		Hashrate:  stats.Pool.Hashrate,
		LastBlock: stats.Pool.LastBlock,
		Miners:    stats.Pool.Miners,
		URL:       stats.Pool.URL,
		Name:      stats.Pool.Name,
	}
	var templateHTML bytes.Buffer
	err = poolTemplate.Execute(&templateHTML, poolData)
	if err != nil {
		log.Fatalf("Unable to load pool template: '%s'", err)
	}
	stats.PoolHTML = templateHTML.String()

	statBytes, err = json.Marshal(&stats)
	if err != nil {
		return "", err
	}
	return string(statBytes), nil
}
