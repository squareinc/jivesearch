$(document).ready(function() {
  // highlight query in the results
  // Highlighting here ensures we don't introduce unsafe characters.
  // This is not ideal and should be replaced with a template 
  // function in Go as this will break if javascript is disabled.
  $(".description").each(function(index, value){
    var q = $("#query").data("query").split(" ");
    var content = $(value).html();
    var c = content.split(" ");
    for (var i = 0; i < c.length; i++){
      for (var j = 0; j < q.length; j++){
        if (c[i].toLowerCase().indexOf(q[j].toLowerCase()) != -1){
          c[i] = "<em>" + c[i] + "</em>";
        }
      }
    }
    $(value).html(c.join(" "));
  });

  // redirect to a default !bang
  $(document).on('click', '.bang_submit', function(){
    changeParamAndRedirect("q", $(this).data('location'));
  });

  // voting
  // TODO: HMAC key
  $(document).on('click', '.arrow', function(){
    var t = $(this);
    var v = $(t).data('vote');
    var removing = false;

    if ($(t).hasClass('voted')){ // already vote for that link?
      removing = true;
      $(t).removeClass('voted');
      v = -1*v;      
    }

    d = {
      'q': $('#query').data('query'), 
      'u': $(t).parent('.vote').data('url'), 
      'v': v
    };
    
    $.ajax({
      type: "POST",
      dataType: "json",
      url: "/vote",
      data: d
    }).done(function(data) {
      $(t).siblings('.arrow').removeClass('voted'); // remove prior vote if it is different
      if (removing != true){
        $(t).addClass('voted');
      }
    }).fail(function(data) {
      $(t).siblings('.arrow').removeClass('voted');
    });
  });

  // Traditional Pagination
  $(document).on('click', '.pagination', function(){
    changeParamAndRedirect("p", $(this).data('page')); 
  });

  function isBang(item) {
    return item.hasOwnProperty("trigger");
  }

  function label(item){
    var label = item.label;
    if (isBang(item)){
      label = "!" + item.trigger;
    }
    return label
  }

  // autocomplete
  $(function(){
    $("#query").autocomplete({
      delay: 25,
      minLength: 1,
      messages: {
        noResults: "",
        results: function() {}
      },
      open: function() {
        $("ul.ui-menu").innerWidth($(this).innerWidth()); // width of input including button
      },
      source: function(request, callback){
        $.getJSON('/autocomplete', {q: request.term}, function(data){ // '{q: request.term}' changes it from ?term=b to ?q=b so nginx doesn't log query.
          callback(data.suggestions);
        });
      },
      select: function(event, ui){
        if (!isBang(ui.item)){ 
          $("#query").val(label(ui.item));
          document.getElementById('form').submit();
          return false;
        }

        // don't submit bang if they aren't done w/ query
        $("#query").val(label(ui.item) + " ");
        return false;
      },
      focus: function(event, ui) {
        $("#query").val(label(ui.item));
        return false;
      },
      }).data('ui-autocomplete')._renderItem = function(ul, item){
        var label = item.label;
        var bang = isBang(item);
        
        if (bang===true){
          label = item.trigger;
        }
        
        // highlight matches in the dropdown
        var re = new RegExp(this.term, 'i');
        var re = new RegExp("^" + this.term);
        var r = label.replace(re, "<span style='font-weight:normal;'>" + "$&" + "</span>");
        var formatted = "<a>" + r + "</a>";

        if (bang===true){
          // Note: below about 600px this doesn't display right.
          // I've tried adding "display:none" for the bang name in main.css but that isn't being respected for some reason.
          r = '<span style="vertical-align:top;">' + item.name + '</span><span style="float:right;margin-right:40%;margin-left:1px;"> !' + r + '</span>';
          formatted = '<a><img width="20" height="20" style="vertical-align:top;" src="' + item.favicon + '"/> ' + r + '</a>';
        }

        return $("<li></li>").data("item.autocomplete", item).append(formatted).appendTo(ul);
      };
  });

  // fix the size of the autocomplete dropdown menu to match the size of the input
  jQuery.ui.autocomplete.prototype._resizeMenu = function(){
    var ul = this.menu.element;
    ul.outerWidth($("#query").outerWidth(true)-40); // 40 is the width of our button
  }

  // redirect "did you mean?" queries
  $("#alternative").on("click", function(){  
    changeParamAndRedirect("q", $(this).attr("data-alternative"));  
  });

  function queryString(){
    return window.location.search;
  }

  var changeParamAndRedirect = function (key, value) {
    var  urlQueryString = document.location.search,
      newParam = key + '=' + value,
      params = '?' + newParam;

    // If the "search" string exists, then build params from it
    if (urlQueryString) {
        var updateRegex = new RegExp('([\?&])' + key + '[^&]*');
        var removeRegex = new RegExp('([\?&])' + key + '=[^&;]+[&;]?');
        if( typeof value == 'undefined' || value == null || value == '' ) { // Remove param if value is empty
            params = urlQueryString.replace(removeRegex, "$1");
            params = params.replace( /[&;]$/, "" );
        } else if (urlQueryString.match(updateRegex) !== null) { // If param exists already, update it
            params = urlQueryString.replace(updateRegex, "$1" + newParam);
        } else { // Otherwise, add it to end of query string
            params = urlQueryString + '&' + newParam;
        }
    }

    // no parameter was set so we don't need the question mark
    params = params == '?' ? '' : params;
    window.location.href = window.location.pathname + params;
  };

  $("#safesearch").show();
  $("#safesearchbtn").on("click", function(){
    $("#safesearch-content").toggle();
  });

  $("#safe").on("click", function(){
    var checked = $("#safe").is(':checked') ? "" : "f";
    changeParamAndRedirect("safe", checked);
  });

  $("#all").on("click", function(){
    // we should delete the param but this works also 
    changeParamAndRedirect("t", "");
  });

  $("#images").on("click", function(){
    changeParamAndRedirect("t", "images");
  });

  $("#maps").on("click", function(){
    changeParamAndRedirect("t", "maps");
  });
});

