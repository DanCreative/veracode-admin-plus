// SubmitRow is run after changing the teams select field
// Gets the form values for the row then makes an HTMX ajax call
function SubmitRow(userId) {
    //console.log("userId: "+userId);
    var trId = "#form-" + userId

    var formValues = htmx.values(htmx.find(trId));
    //console.log("Form Values: " + formValues);
    htmx.ajax('PUT', '/cart/users/' + userId, {
        values: formValues
    });
}

// AlterRow is run after changing any of the selects or inputs in the row
// Enables the cart buttons and adds a CSS class to the row to indicate that changes have been made
function AlterRow(trId) {
    var tr = $("#" + trId);
    //tr.trigger("submit");
    tr.addClass('altered');

    $(".cart-controls").attr("disabled", false);
    //console.log("Triggered submit on " + trId);
}

// CascadeValues is run after the filter by select changes
// It shows the available values in the #filter-options select depending
// on the value chosen in the filter by select
function CascadeValues(name) {
    //console.log("Triggered CascadeValues with " + name);
    $("#filter-options").attr("name", name)
    $("#filter-options").children().each(function () {
        if ($(this).hasClass(name)) {
            $(this).removeAttr("hidden");
            //console.log("Removing hidden from " + this);
        } else {
            $(this).attr("hidden", true);
            //console.log("Adding hidden to " + this);
        }
    });
}

// ShowAdminTeams runs after changing the teams select
// It updates the teams available in the admin teams select
function ShowAdminTeams(sourceId, targetId) {
    var teamsSelectVals = $("#" + sourceId).val();
    var admTeamsSelect = $("#" + targetId);
    admTeamsSelect.children().each(function () {
        if (teamsSelectVals.length < 1) {
            $(this).attr("hidden", true);
        } else {
            for (let index = 0; index < teamsSelectVals.length; index++) {
                if (this.value == teamsSelectVals[index]) {
                    $(this).removeAttr("hidden");
                }
            }
        }
    });
}

// TeamAdminModalCheckActive runs after the teamadmin role checkbox is clicked
// It enables/disables the button to open the team admin modal depending on
// whether the checkbox has been checked or unchecked
function TeamAdminModalCheckActive(admteamSelectId, element) {
    var checkbox = $(element);
    var button = checkbox.next();

    if (checkbox.is(":checked")) {
        button.addClass("active");
    } else {
        button.removeClass("active");
        $("#" + admteamSelectId).val([]);
    }
}

// TeamAdminModal is run after the the teams admin button is clicked
// It makes sure that the checkbox is actually checked then it
// opens the team admin modal
function TeamAdminModal(modalId, element) {
    var checkbox = $(element).prev();
    //console.log(checkbox);
    if (checkbox.is(":checked")) {
        //console.log(checkbox.attr('id') + " is checked");
        ShowModal(modalId);
    } else {
        console.log("Not opening team admin modal, " + checkbox.attr('id') + " is not checked");
    }
}

// ShowModal is a helper function that opens any modal using an id
function ShowModal(id) {
    //console.log("Trying to find modal: " + id);
    var modal = $("#" + id);
    modal.show();
    //modal.style.display = "block";
}

// CloseAnyModal is a helper function that closes any of the modals
function CloseAnyModal(event) {
    var modals = document.getElementsByClassName("modal");
    for (let index = 0; index < modals.length; index++) {
        const modal = modals[index];
        if (event.target == modal) {
            modal.style.display = "none";
        }
    }
}

// CloseParentModal is run when the X button in a modal is clicked
// It closes the modal
function CloseParentModal(el) {
    $(el).closest(".modal").hide();
}

// CheckScanTypesEmpty sets the any scan type to true if non of the 
// other scan types are checked. This is done to avoid the issue where
// a user can be submitted to the backend without any scan types while
// a role that requires atleast one is selected.
function CheckScanTypesEmpty(userId) {
    var isEmpty = true
    $(".scan-type.roles-" + userId).each(function () {
        if ($(this).is(":checked")) {
            isEmpty = false
        }
    });

    //console.log("CheckSanTypesEmpty() -> isEmpty = "+ isEmpty);
    if (isEmpty) {
        $("#extsubmitanyscan-" + userId).prop("checked", true);
    }
}

// AnyScanClicked is run when the Any Scan role is selected
// It clears all of the other scan type roles
// If all of the roles are deselected, any scan role will be re-selected
function AnyScanClicked(el, userId) {
    if ($(el).is(":checked")) {
        $(".not-any-scan.roles-" + userId).each(function () {
            $(this).val([]);
        });
    } else {
        CheckScanTypesEmpty(userId);
    }
}

// NanyScanClicked is run when any of the Non-Any Scan roles are selected
// It clears the Any Scan roles
// If all of the roles are deselected, any scan role will be re-selected
function NanyScanClicked(el, userId) {
    if ($(el).is(":checked")) {
        $("#extsubmitanyscan-" + userId).val([]);
    } else {
        CheckScanTypesEmpty(userId);
    }
}

// CloseDropdowns closes all dropdowns except for the dropdown that was triggered
// by the button.
function CloseDropdowns(event) {
    var target = $(event.target);

    $(".dropdown").each(function () {
        if (!target.is($(this).prev(".action-button"))) {
            $(this).removeClass("show")
        }
    });
}

function ShowActionDropdown(event) {
    $(event.target).next("div").addClass("show");
}

window.onclick = function (event) {
    CloseAnyModal(event);
    CloseDropdowns(event);
}