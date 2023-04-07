package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// IPInterfaceGetDataModelONTAP describes the GET record data model using go types for mapping.
type IPInterfaceGetDataModelONTAP struct {
	Name    string `mapstructure:"name"`
	Scope   string `mapstructure:"scope"`
	SVMName string `mapstructure:"svm.name"`
	UUID    string `mapstructure:"uuid"`
}

// IPInterfaceResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type IPInterfaceResourceBodyDataModelONTAP struct {
	Name     string                      `mapstructure:"name"`
	SVM      Vserver                     `mapstructure:"svm"`
	IP       IPInterfaceResourceIP       `mapstructure:"ip"`
	Location IPInterfaceResourceLocation `mapstructure:"location"`
}

// IPInterfaceResourceIP is the body data model for IP field
type IPInterfaceResourceIP struct {
	Address string `mapstructure:"address"`
	Netmask int64  `mapstructure:"netmask"`
}

// IPInterfaceResourceLocation is the body data model for location field
type IPInterfaceResourceLocation struct {
	HomeNode *IPInterfaceResourceHomeNode `mapstructure:"home_node,omitempty"`
	HomePort *IPInterfaceResourceHomePort `mapstructure:"home_port,omitempty"`
}

// IPInterfaceResourceHomeNode is the body data model for home_node field
type IPInterfaceResourceHomeNode struct {
	Name string `mapstructure:"name"`
}

// IPInterfaceResourceHomePort is the body data model for home_port field
type IPInterfaceResourceHomePort struct {
	Name string                      `mapstructure:"name"`
	Node IPInterfaceResourceHomeNode `mapstructure:"node"`
}

// GetIPInterface to get ip_interface info
func GetIPInterface(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*IPInterfaceGetDataModelONTAP, error) {
	api := "network/ip/interfaces"
	query := r.NewQuery()
	query.Set("name", name)
	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	query.Fields([]string{"name", "svm.name", "ip", "scope"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading ip_interface info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP IPInterfaceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read ip_interface data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetIPInterfaces to get ip_interface info for all resources matching a filter
func GetIPInterfaces(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *IPInterfaceGetDataModelONTAP) ([]IPInterfaceGetDataModelONTAP, error) {
	api := "network/ip/interfaces"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "ip", "scope"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding ip_interface filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading ip_interfaces info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []IPInterfaceGetDataModelONTAP
	for _, info := range response {
		var record IPInterfaceGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read ip_interface data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateIPInterface to create ip_interface
func CreateIPInterface(errorHandler *utils.ErrorHandler, r restclient.RestClient, body IPInterfaceResourceBodyDataModelONTAP) (*IPInterfaceGetDataModelONTAP, error) {
	api := "network/ip/interfaces"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding ip_interface body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating ip_interface", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP IPInterfaceGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding ip_interface info", fmt.Sprintf("error on decode storage/ip_interfaces info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create ip_interface source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteIPInterface to delete ip_interface
func DeleteIPInterface(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "network/ip/interfaces"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting ip_interface", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}