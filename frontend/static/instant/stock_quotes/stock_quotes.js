$(document).ready(function() {
    $("#answer").width("630px").height("385px");

    var margin = {top: 20, right: 70, bottom: 30, left: 50},
        width = 530 - margin.left - margin.right, // was 960
        height = 180 - margin.top - margin.bottom; // was 500

    var parseTime = d3.timeParse("%Y-%m-%dT%H:%M:%S%Z"); // 2018-01-12T00:00:00Z
    var formatTime = d3.timeFormat("%a %B %e, %Y");
    var formatNumber = d3.format(",");
    var bisectDate = d3.bisector(function(d) { 
        return d.date; 
    }).left;

    var x = d3.scaleTime().range([0, width]);
    var y = d3.scaleLinear().range([height, 0]);
    var y1 = d3.scaleLinear().range([height, 0]);

    var line = d3.line().x(function(d){ return x(d.date); }).y(function(d){ return y(d.close); });
    var area = d3.area().x(function(d){ return x(d.date); }).y1(function(d){ return y(d.close); });
    var div = d3.select("#stock_chart").append("div").attr("class", "tooltip").style("opacity", 0);
    var svg = d3.select("#stock_chart").append("svg").attr("width", width + margin.left + margin.right)
                .attr("height", height + margin.top + margin.bottom)
                .append("g").attr("transform","translate(" + margin.left + "," + margin.top + ")");

    data.forEach(function(d) {
        d.date = parseTime(d.date);
        d.close = +d.close;
    });

    x.domain(d3.extent(data, function(d) { return d.date; }));
    y.domain([0, d3.max(data, function(d) { return d.close; })]);
    y1.domain([0, d3.max(data, function(d) { return d.close; })]);
    area.y0(y(0));

    svg.append("path").datum(data).attr("class", "line").attr("d", area);
    svg.append("g").call(d3.axisBottom(x).ticks(5).tickFormat(d3.timeFormat("%Y"))).attr("class", "axis").attr("transform", "translate(0," + height + ")");
    svg.append("g").call(d3.axisLeft(y).ticks(5)).attr("class", "axis");
    svg.append("g").call(d3.axisRight(y1)).attr("class", "rightAxis").attr("transform", "translate( " + width + ", 0 )");

    var focus = svg.append("g").attr("class", "focus").style("display", "none");
    focus.append("line").attr("class", "x-hover-line hover-line").attr("y1", 0).attr("y2", height);
    focus.append("line").attr("class", "y-hover-line hover-line").attr("x1", width).attr("x2", width);
    focus.append("circle").attr("r", 1.5);
    focus.append("text").attr("x", 15).attr("dy", ".31em");

    svg.append("rect")
        .attr("transform", "translate(" + 1 + "," + 1 + ")")
        .attr("class", "overlay").attr("width", width).attr("height", height)
        .on("mouseover", function(d) { 
            focus.style("display", null);
        })
        .on("mouseout", function() { 
            div.style("opacity", 0);
            focus.style("display", "none"); 
        })
        .on("mousemove", mousemove);

    function mousemove() {
        var x0 = x.invert(d3.mouse(this)[0]),
            i = bisectDate(data, x0, 1),
            d0 = data[i - 1],
            d1 = data[i],
            d = x0 - d0.date > d1.date - x0 ? d1 : d0;
            focus.attr("transform", "translate(" + x(d.date) + "," + y(d.close) + ")");
            focus.select(".x-hover-line").attr("y2", height - y(d.close));
            focus.select(".y-hover-line").attr("x2", width + width);
            focus.select("text").text(function() { 
                div.html("<em>" + formatNumber(d.close) + "</em> " + formatTime(d.date))
                    .style("left", (d3.event.pageX) + "px").style("top", (d3.event.pageY - 28) + "px");
                div.style("opacity", .9);
                return;
            });        
    }
});
