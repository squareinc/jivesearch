window.onload = function() {
    var inputs = $(".unit > input");
    var selects = $(".unit > select");

    $(inputs).each(function(i) {
        $(this).on('input', function(){
            if (!isNaN(inputs[i].value)){
                // change the opposite input value
                var j = ((i===0) ? 1 : 0);
                inputs[j].value = convert(i);
            }
        });
    });

    $(selects).each(function(i) {
        // Selecting a different unit should ALWAYS
        // change the second input box. Never the first.
        $(this).change(function() {
            if (!isNaN(inputs[i].value)){
                inputs[1].value = convert(0);
            }
        });
    });

    function convert(i){
        var val = inputs[i].value; // value we are converting FROM
        var rate = selects[i].value; // units we are converting FROM
        var j = ((i===0) ? 1 : 0);
        var rate2 = selects[j].value; // units we are converting TO
        return (val * rate) / rate2;
    }

    if (!isNaN(inputs[0].value)){
        inputs[1].value = convert(0);
    }
}        
