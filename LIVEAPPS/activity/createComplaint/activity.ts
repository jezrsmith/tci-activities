/*
 * Copyright © 2017. TIBCO Software Inc.
 * This file is subject to the license terms contained
 * in the license file that is distributed with this file.
 */
import {Observable} from "rxjs/Observable";
import {Injectable, Injector, Inject} from "@angular/core";
import {Http} from "@angular/http";
import {
    WiContrib,
    WiServiceHandlerContribution,
    IValidationResult,
    ValidationResult,
    IFieldDefinition,
    IActivityContribution,
    IConnectorContribution,
    WiContributionUtils
} from "wi-studio/app/contrib/wi-contrib";

@WiContrib({})
@Injectable()
export class CreateComplaintActivityContributionHandler extends WiServiceHandlerContribution {
    constructor( @Inject(Injector) injector, private http: Http) {
        super(injector, http);
    }

// this gets the list of connections defined for live apps so the user can select one from the box
    value = (fieldName: string, context: IActivityContribution): Observable<any> | any => {
        if (fieldName === "liveappsConnection") {
            return Observable.create(observer => {
            let connectionRefs = [];
            // NOTE: LIVEAPPS is the connection CATEGORY from the connection's connector.json
            WiContributionUtils.getConnections(this.http, "LIVEAPPS").subscribe((data: IConnectorContribution[]) => {
                data.forEach(connection => {
                    /**
                     * Create a list with all LIVEAPPS connectors that have been created by the user 
                     */
                    for (let i = 0; i < connection.settings.length; i++) {
                        if (connection.settings[i].name === "name") {
                            connectionRefs.push({
                                "unique_id": WiContributionUtils.getUniqueId(connection),
                                "name": connection.settings[i].value
                            });
                            break;
                        }
                    }
                });
                observer.next(connectionRefs);
            });
        });

        }
        return null;
    }

    // This makes sure a value is selected for the connection
    validate = (fieldName: string, context: IActivityContribution): Observable<IValidationResult> | IValidationResult => {
        if (fieldName === "liveappsConnection") {
            let connection: IFieldDefinition = context.getField("liveappsConnection")
            if (connection.value === null) {
                return ValidationResult.newValidationResult().setError("LIVEAPPS-MSG-1000", "Live Apps Connection must be configured");
            }
        }
        return null;
    }

}