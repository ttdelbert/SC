package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)


type PrivateAsset struct {
	ObjectType string `json:"docType"`	//docType is used to distinguish the various types of objects in state database
	AssetName string	`json:"assetname"`
	AssetProperty string	`json:"assetproperty"`
	Owner string	`json:"owner"`
	Price int `json:"price"`
}

type SmartContract struct {
	contractapi.Contract
}

//PrivateDeal creates a private transaction between original owner and dealer
func (s *SmartContract) PrivateDeal (ctx contractapi.TransactionContextInterface) error {
	transMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting transient" + err.Error())
	}
	//Get passed in transient field----must be scprivasset:...
	transientAssetJSON, ok := transMap["scprivasset"]
	if !ok {
		return fmt.Errorf("Private data not found in the transient map")
	}

	type scassetTransientInput struct {
		AssetName string	`json:"assetname"`
		AssetProperty string 	`json:"assetproperty"`
		Owner string	`json:"owner"`
		Price int 	`json:"price"`
	}

	var scassetinput scassetTransientInput
	err = json.Unmarshal(transientAssetJSON, &scassetinput)

	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}
	if len(scassetinput.AssetName) == 0 {
		return fmt.Errorf("assetname field must be a non-empty string")
	}
	if len(scassetinput.AssetProperty) == 0 {
		return fmt.Errorf("assetproperty field must be a non-empty string")
	}
	if len(scassetinput.Owner) == 0 {
		return fmt.Errorf("owner field must be a non-empty string")
	}
	if scassetinput.Price <= 0 {
		return fmt.Errorf("price field must be a positive integer")
	}

	//check if scpriveasset already exists
	scprivasset, err := ctx.GetStub().GetPrivateData("collectionPrivateData", scassetinput.AssetName)
	if err != nil {
		return fmt.Errorf("failed to get scprivasset: " + err.Error())
	}else if scprivasset != nil {
		fmt.Printf("This scprivasset already exists: " + scassetinput.AssetName)
		return fmt.Errorf("This scprivasset already exists: " + scassetinput.AssetName)
	}

	//Create assetprivate object, marshal to JSON, save to state
	assetprivate := &PrivateAsset{
		ObjectType: "SCPrivAsset",
		AssetName: scassetinput.AssetName,
		AssetProperty: scassetinput.AssetProperty,
		Owner: scassetinput.Owner,
		Price: scassetinput.Price,
	}

	assetprivateJSON, err := json.Marshal(assetprivate)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	//Save assetprivate to state
	err = ctx.GetStub().PutPrivateData("collectionPrivateData", scassetinput.AssetName, assetprivateJSON)
	if err != nil {
		return fmt.Errorf("failed to put assetprivate: %s", err.Error())
	}

	return nil
}

//ReadPrivateAsset Reads private asset by assetname
func (s *SmartContract) ReadAssetPrivate(ctx contractapi.TransactionContextInterface, assetname string) (*PrivateAsset, error){
	assetprivateJSON, err := ctx.GetStub().GetPrivateData("collectionPrivateData", assetname)
	if err != nil {
		return nil, fmt.Errorf("failed to read private asset: %s", err.Error())
	}
	if assetprivateJSON == nil {
		return nil, fmt.Errorf("%s does not exist", assetname)
	}
	assetprivate := new(PrivateAsset)

	err = json.Unmarshal(assetprivateJSON, assetprivate)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %s", err.Error())
	}

	return assetprivate, nil
}

//GetAllPrivateAssets returns all privateassets
func (s *SmartContract) GetAllPrivateAssets(ctx contractapi.TransactionContextInterface)([]*PrivateAsset, error){
	resultIterator, err := ctx.GetStub().GetPrivateDataByRange("collectionPrivateData", "", "")
	if err != nil {
		return nil, err
	}

	defer resultIterator.Close()

	var privateassets []*PrivateAsset
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return nil, err
		}

		var privateasset PrivateAsset
		err = json.Unmarshal(queryResponse.Value, &privateasset)
		if err != nil {
			return nil, err
		}
		privateassets = append(privateassets, &privateasset)
	}
	return privateassets, nil
}

//GetPrivateAssetHash provides a func to verify an asset existing or not
func (s *SmartContract) GetPrivateAssetHash(ctx contractapi.TransactionContextInterface, assetname string) (string, error) {
	hashAsbytes, err := ctx.GetStub().GetPrivateDataHash("collectionPrivateData", assetname)
	if err != nil {
		return "", fmt.Errorf("failed to get hash for privateasset" + err.Error())
	}else if hashAsbytes == nil {
		return "", fmt.Errorf("Asset does not exist" + assetname)
	}

	return string(hashAsbytes), nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create private asset chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting private asset chaincode: %s", err.Error())
	}
}