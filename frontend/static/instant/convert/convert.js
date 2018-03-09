window.onload = function() {
    var inputs = $(".digital_storage > input");
    var selects = $(".digital_storage > select");
    
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
        var val = inputs[i].value;
        var units = selects[i].value; // units we are converting FROM
        
        var j = ((i===0) ? 1 : 0);
        var units2 = selects[j].value; // units we are converting TO

        return (val * units) / units2;
    }
}