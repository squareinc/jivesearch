$(document).ready(function() {
  $.ajax({url: "/answer?q=$MSFT&o=json", success: function(result){
    //$("#answer").html(result);
  }});
});