// A common pattern used for submitting data/forms on this website
// It requires the html elements to be set up in a certain way
// returns the request so custom handlers can be added
function SubmitForm(url, formId, buttonId, alertId, buttonInit = "Submit", buttonProcessing = "Submitting", buttonSuccess = "Success",){
    // Client side validator from jquery plugin
    if (!$(formId).valid())
            return

    $(buttonId).html('<span class="spinner-border spinner-border-sm"></span>'+buttonProcessing)
    $(alertId).collapse('hide')
    return $.ajax({
        type: 'POST',
        url: url,
        data: $(formId).serialize(),
    }).done(function () {
        $(buttonId).html(buttonSuccess)
    }).fail(function (xhr, status, error) {
        $(alertId).text(xhr.responseText)
        $(alertId).collapse('show')
        $(buttonId).html(buttonInit)
    });
}

function FormToMap(formId){
    let fd = new FormData($(formId)[0]);
    let data = {};
    for (let [key, prop] of fd) {
        if (!isNaN(parseFloat(prop)))
            data[key] = parseFloat(prop);
        else
            data[key] = prop;
    }
    return data
}


// Enables easy tooltips everywhere
$(function () {
    $('[data-toggle="tooltip"]').tooltip()
  })


// Custom selector for bootstrap that will toggle the next element in DOM
$(document).on('click', '.collapse-next', function() {
    $(this).next().collapse('toggle')
});

// General filtering method
$(document).on('keyup', 'input.filter', function(){
    var target = $(this).data('target');
    var text = $(this).val().toLowerCase();
    $(target).filter(function() {$(this).toggle($(this).text().toLowerCase().indexOf(text) > -1);});
});


function round(num, prec){
    s = Math.pow(10, prec)
    return Math.round(num * s) / s
}