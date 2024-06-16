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
function CloseParentModal(event) {
    $(event.target).closest(".modal").remove();
}

function ShowScanTypesModal(event) {
    console.log($(event.target).parent().next("div"));
    $(event.target).closest("div").next("div").show();
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

function PutUser(event) {
    var userRow = $(event.target).closest('tr')
    userRow.find(".user-submit-button").attr("disabled", false);
    // userRow.find(".user-clear-button").attr("disabled", false);
    userRow.addClass('altered');
    //$(".cart-controls").attr("disabled", false);
}

function SpecificScanTypesSelected(event) {
    var parent = $(event.target).parent()
    // Disable the anyscan input
    parent.prev().find("#any-scan-role").attr("disabled", true);

    // Re-enable the not-any scan inputs
    parent.find("input.nany-scan-role").each(function () {
        $(this).attr("disabled", false);
    });
    $(event.target).parent().children("ul").removeClass("hide");
    console.log($(event.target).parent().next());
}

function AnyScanTypesSelected(event) {
    // Re-enable the anyscan input
    $(event.target).next().attr("disabled", false);

    // Disable the not-any scan inputs
    $(event.target).parent().next().find("input.nany-scan-role").each(function () {
        $(this).attr("disabled", true);
    });

    $(event.target).parent().next().children("ul").addClass("hide");
}