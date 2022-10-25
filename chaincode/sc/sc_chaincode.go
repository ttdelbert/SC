package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
)

type Asset struct {
	AssetName string	`json:"assetname"`
	AssetProperty string	`json:"assetproperty"`
	Owner string	`json:"owner"`
	AssetProcessInfo string	`json:"assetprocessinfo"`
	AssetOriginalPrice int `json:"assetoriginalprice"`
	AssetProcessPrice int	`json:"assetprocessprice"`
}

type SmartContract struct {
	contractapi.Contract
}

//AssetExists returns true when asset with given assetname exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, assetname string) (bool, error){
	assetJSON, err := ctx.GetStub().GetState(assetname)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return assetJSON != nil, nil
}

//CreateAsset issues a brand-new asset to the world state by first owner
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetname string, assetproperty string, owner string) error {
	exist, err := s.AssetExists(ctx, assetname)
	if err != nil {
		return err
	}
	if exist {
		return fmt.Errorf("the asset %s already exists", assetname)
	}

	asset := Asset{
		AssetName: assetname,
		AssetProperty: assetproperty,
		Owner: owner,
		AssetProcessInfo: "",
		AssetOriginalPrice: 0,
		AssetProcessPrice: 0,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetname, assetJSON)
}

//ReadAsset returns the asset stored in the world state with given assetname
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, assetname string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(assetname)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", assetname)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

//SellOriginalAsset updates the owner field and fills the assetoriginalprice field as selling the original asset to processing industry or distributor
func (s *SmartContract) SellOriginalAsset (ctx contractapi.TransactionContextInterface, assetname string, newowner string, assetoriginalprice int) error {
	asset, err := s.ReadAsset(ctx, assetname)
	if err != nil {
		return err
	}

	asset.Owner = newowner
	asset.AssetOriginalPrice = assetoriginalprice

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetname, assetJSON)
}

//ProcessAsset fills the assetprocessinfo field as proceccing industry or distributor processes the original asset
func(s *SmartContract) ProcessAsset (ctx contractapi.TransactionContextInterface, assetname string, assetprocessinfo string) error {
	asset, err := s.ReadAsset(ctx, assetname)
	if err != nil {
		return err
	}

	asset.AssetProcessInfo = assetprocessinfo
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetname, assetJSON)
}

//SellProcessedAsset updates the owner field and fills the assetprocessprice field as selling the processed asset to dealer
func (s *SmartContract) SellProcessedAsset (ctx contractapi.TransactionContextInterface, assetname string, newowner string, assetprocessprice int) error {
	asset, err := s.ReadAsset(ctx, assetname)
	if err != nil {
		return err
	}

	asset.Owner = newowner
	asset.AssetProcessPrice = assetprocessprice

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetname, assetJSON)
}

//GetallAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets (ctx contractapi.TransactionContextInterface) ([]*Asset, error){
	resultIterator, err := ctx.GetStub().GetStateByRange("","")
	if err != nil {
		return nil, err
	}

	defer resultIterator.Close()

	var assets []*Asset
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}
	return assets, nil
}

//GetHistory returns the operational history of asset by given assetname
func (s *SmartContract) GetHistory (ctx contractapi.TransactionContextInterface, assetname string) ([]string, error ){
	exist, err := s.AssetExists(ctx, assetname)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, fmt.Errorf("the asset %s does not exist", assetname)
	}

	resultIterator, err := ctx.GetStub().GetHistoryForKey(assetname)
	if err != nil {
		return nil, err
	}

	defer resultIterator.Close()

	var assethistory []string
	for resultIterator.HasNext() {
		response, err := resultIterator.Next()
		if err != nil {
			return nil, err
		}

		txvalue := response.Value
		txtimesamp := response.Timestamp
		tm := time.Unix(txtimesamp.Seconds, 0)
		datastr := tm.Format("2021-09-12 03:04:05 PM")
		assethistory = append(assethistory, datastr + ":" + string(txvalue))
	}

	return assethistory, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create supply chain chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting supply chain chaincode: %s", err.Error())
	}
}



