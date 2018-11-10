$(document).ready(function() {
  // highlight query in the results
  // Highlighting here ensures we don't introduce unsafe characters.
  // This is not ideal and should be replaced with a template 
  // function in Go as this will break if javascript is disabled.
  var highlight = function(value){
    var q = String($("#query").data("query")).split(" ");
    var content = $(value).html();
    var c = content.split(" ");
    for (var i = 0; i < c.length; i++){
      for (var j = 0; j < q.length; j++){
        if (c[i].toLowerCase().indexOf(q[j].toLowerCase()) != -1){
          c[i] = "<em>" + c[i] + "</em>";
        }
      }
    }
    return c.join(" ")
  }

  // this is a workaround for https://github.com/jivesearch/jivesearch/issues/66
  if ($(".document").length === 0){
    $("#empty").hide();
    fetch(1);
  };

  $(".description").each(function(index, value){
    $(value).html(highlight(value));
  });

  // redirect to a default !bang
  $(document).on('click', '.bang_submit', function(){
    params = changeParam("q", $(this).data('location'));
    redirect(params);
  });

  // Traditional Pagination
  $(document).on('click', '.pagination', function(){
    params = changeParam("p", $(this).data('page'));
    redirect(params);
  });

  // Infinite Scroll
  // ===== Scroll to Top ==== 
  $(window).scroll(function() {
    if ($(this).scrollTop() >= 400) {     // If page is scrolled more than 400px
      $('#return-to-top').fadeIn(200);    // Fade in the arrow
    } else {
      $('#return-to-top').fadeOut(200);   // Else fade out the arrow
    }
  });
  $('#return-to-top').click(function() {  
    $('body,html').animate({
      scrollTop : 0                       
    }, 500);
  });
  
  var fetching = false;
  $(window).scroll(function() {
    if (($("#infinite_scroll").length == 1) && (fetching===false) && ($(window).scrollTop() == ($(document).height() - $(window).height())) - 25) {
      fetching = true;
      $("#loading").show();
      var page = $("#next_page").attr("data-page");
      if (page === undefined){
        return;
      }

      fetch(page);
    }
  }); 

  function fetch(page){
    var params = changeParam("p", page);
    params = params + "&o=json"; // add the new param
    var u = window.location.pathname + params;
    $.ajax({
      url: u,
    }).done(function(data) {
      $("#next_page").attr("data-page", data.search.next);
      var i;
      for (i = 0; i < data.search.documents.length; i++) {
        /*
        This is a workaround for empty search results we get sometimes...
        can't simply clone as we may not have results for first page.
        */
        var doc = data.search.documents[i];
        var desc = highlight(`<div class="description">`+doc.description+`</div>`); // bit redundant to repeat the <div tag here...
        var h = `<div class="pure-u-1">
          <div class="pure-u-22-24 pure-u-md-21-24 result">
            <div class="title"><a href="`+doc.id+`" rel="noopener">`+doc.title+`</a></div>
            <div class="url">`+doc.id.substring(0,80)+`</div>
            <div class="description">`+desc+`</div>
          </div>
        </div>`;

        $("#documents").append(h);
      }
      fetching = false;
    }).always(function(data) {
      $("#loading").hide();
    });
  }

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

  // Wikipedia disambiguation page link & other links w/in Wikipedia snippets
  $(document).on('click', '.wikipedia_disambiguation, .wikipedia_item', function(){
    params = changeParam("q", $(this).data('title'));
    redirect(params);
  });
  
  // redirect "did you mean?" queries
  $("#alternative").on("click", function(){  
    params = changeParam("q", $(this).attr("data-alternative")); 
    redirect(params);
  });

  $("#safesearch").show();
  $("#safesearchbtn").on("click", function(){
    $("#safesearch-content").toggle();
  });

  $("#safe").on("click", function(){
    var checked = $("#safe").is(':checked') ? "" : "f";
    params = changeParam("safe", checked);
    redirect(params);
  });

  $("#search_filter").on('change', function() {
    var checked = $('input[name=search_filter]:checked', '#search_filter').val();
    params = changeParam("f", checked);
    redirect(params);
  });

  $("#all").on("click", function(){
    // we should delete the param but this works also 
    params = changeParam("t", "");
    redirect(params);
  });

  $("#images").on("click", function(){
    params = changeParam("t", "images");
    redirect(params);
  });

  $("#map, #maps").on("click", function(){
    params = changeParam("t", "maps");
    redirect(params);
  });

});