/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/
jQuery(function(){
	
	new AjaxUtil({
		url: "/api/users/current",
		type: "get",
		success: function(data, status, xhr){},
		errors: {
			403: ""
		},
		error: function(jqXhr){
			if(jqXhr && jqXhr.status == 401){
				document.location = "/signIn";				
		    }
		}
	}).exec();
	
	$("#divErrMsg").css({"display": "none"});
	
	$("#OldPassword,#Password,#ConfirmedPassword").on("blur", validateCallback);
	validateOptions.Items = ["#OldPassword", "#Password", "#ConfirmedPassword"];
	
	/*
	function bindEnterKey(){
		$(document).on("keydown", function(e){
			if(e.keyCode == 13){
			  e.preventDefault();
			  if($("#txtCommonSearch").is(":focus")){
				document.location = "/search?q=" + $("#txtCommonSearch").val();
			  }else{
			    $("#btnSubmit").trigger("click");	
			  }
			}
		});
	}
	function unbindEnterKey(){
		$(document).off("keydown");
	}
	bindEnterKey();
*/
	keydownNS.bindEnterKey();
	
	var spinner = new Spinner({scale:1}).spin();

	$("#btnSubmit").on("click", function(){
		validateOptions.Validate(function(){
			var oldPassword = $("#OldPassword").val();
			var password = $("#Password").val();
			$.ajax({
				"url": "/updatePassword",
				"type": "post",
				"data": {"old_password": oldPassword, "password" : password},
				"beforeSend": function(e){
				   keydownNS.unbindEnterKey();
				   $("h1").append(spinner.el);
				   $("#btnSubmit").prop("disabled", true);	
				},
				"success": function(data, status, xhr){
					if(xhr && xhr.status == 200){
						$("#dlgModal")
							.dialogModal({
								"title": i18n.getMessage("title_change_password"), 
								"content": i18n.getMessage("change_password_successfully"),
								"callback": function(){ 								
									window.close();
								}
							});
					}
				},
				"error": function(jqXhr, status, error){
					$("#dlgModal")
						.dialogModal({
							"title": i18n.getMessage("title_change_password"), 
							"content": i18n.getMessage(jqXhr.responseText),
							"callback": function(){ 
								keydownNS.bindEnterKey();
								return;
							}
						});
				},
				"complete": function(){
					spinner.stop();
					$("#btnSubmit").prop("disabled", false);	
				}
			});
		});
	});
});
