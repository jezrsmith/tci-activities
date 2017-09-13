/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */

import {Injectable} from "@angular/core";
import {WiContrib, WiServiceHandlerContribution, AUTHENTICATION_TYPE} from "wi-studio/app/contrib/wi-contrib";
import {IConnectorContribution, IFieldDefinition, IActionResult, ActionResult} from "wi-studio/common/models/contrib";
import {Observable} from "rxjs/Observable";
import {IValidationResult, ValidationResult, ValidationError} from "wi-studio/common/models/validation";

@WiContrib({})
@Injectable()
export class TibcoLiveappsConnectorContribution extends WiServiceHandlerContribution {
    constructor() {
        super();
    }

   
    value = (fieldName: string, context: IConnectorContribution): Observable<any> | any => {
		return null;
    }
 
    validate = (name: string, context: IConnectorContribution): Observable<IValidationResult> | IValidationResult => {
		if( name === "Save") {
         let username: IFieldDefinition;
         let password: IFieldDefinition;
         let region: IFieldDefinition;
         
         for (let configuration of context.settings) {
    		if( configuration.name === "username") {
    		   username = configuration
    		} else if( configuration.name === "password") {
    		   password = configuration
    		} else if( configuration.name === "region") {
    		   region = configuration
    		}
		 }
		 
         if(username.value && password.value && region.value) {
            // Enable Connect button
            return ValidationResult.newValidationResult().setReadOnly(false)
         } else {
            return ValidationResult.newValidationResult().setReadOnly(true)
         }
      }
       return null;
    }

    action = (actionName: string, context: IConnectorContribution): Observable<IActionResult> | IActionResult => {
		
		if( actionName == "Save") {
          return Observable.create(observer => {
         	let username: IFieldDefinition;
         	let password: IFieldDefinition;
         	let region: IFieldDefinition;
         
         	for (let configuration of context.settings) {
    			if( configuration.name === "username") {
    		   		username = configuration;
    			} else if( configuration.name === "password") {
    		   		password = configuration;
    			} else if( configuration.name === "region") {
    		   		region = configuration;
    			}
		 	}
			
			 // should test the connection properly here
			 // for now we will just set the data

			let actionResult = {
				context: context,
				authType: AUTHENTICATION_TYPE.BASIC,
				authData: {}
			}			
				
				/**
                 * Call the observer and tell it the validation was sucessful and the data should be saved
                 */
				 observer.next(ActionResult.newActionResult().setSuccess(true).setResult(actionResult));
		 });
       }
       return null;
    }
}