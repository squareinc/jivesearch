var changeParam = function(key, value) {
  var  urlQueryString = document.location.search,
       newParam = key + '=' + value,
       params = '?' + newParam;

  // If the "search" string exists, then build params from it
  if (urlQueryString) {
      var updateRegex = new RegExp('([\?&])' + key + '[^&]*');
      if( typeof value == 'undefined' || value == null || value == '' ) { // Remove param if value is empty
          params = removeParam(key);
      } else if (urlQueryString.match(updateRegex) !== null) { // If param exists already, update it
          params = urlQueryString.replace(updateRegex, "$1" + newParam);
      } else { // Otherwise, add it to end of query string
          params = urlQueryString + '&' + newParam;
      }
  }

  // no parameter was set so we don't need the question mark
  params = params == '?' ? '' : params;
  return params
};

var removeParam = function(key){
  var  urlQueryString = document.location.search;
  var removeRegex = new RegExp('([\?&])' + key + '=[^&;]+[&;]?');
  params = urlQueryString.replace(removeRegex, "$1");
  params = params.replace( /[&;]$/, "" );
  return params;
};

var redirect = function(params){
  window.location.href = window.location.pathname + params;
};