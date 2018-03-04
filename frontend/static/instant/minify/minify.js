$(document).ready(function() {
    $("#minify,#prettify").on("click", function(){
        if (!$("#code_input").val()){
            return
        }

        var options = {
            source: $("#code_input").val(),
            mode: $(this).data("value"),
            lang: "auto",
            inchar: "\t",
            insize: 1
        };
        
        $("#code_output").val(global.prettydiff.prettydiff(options));
        $("#code_output").show();        
    });
});