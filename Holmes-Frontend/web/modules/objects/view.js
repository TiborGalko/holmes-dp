'use strict';

function get_all_objects(){
    console.log("get_object called");
    $.ajax({
        type: 'POST',
        url: current_env.get('api_url'),
        processData: false,
        contentType: 'application/json',
        data: JSON.stringify({
            module: "objects",
            action: "getAll",
        }),
        success: function(r) {
            if(r.error != ""){
                console.log("success but error");
                $.growl.warning({ title: "An error occured!", message: r.error, size: 'large' });
            } else {
                let tbody =  $('#object-table > tbody');
                tbody.empty();
                let cnt = 1;
                let sha256 = "";
                $.each(r.result.Object, function (key, object) {
                    //console.log(object);
                    tbody.append('<tr>');
                    tbody.append("<td>" + cnt + "</td>");
                    cnt++;

                    $.each(object, function(k, v){
                        /*console.log(k);
                        console.log(v);*/
                        if (k === "sha256") {
                            sha256 = v;
                        }
                        tbody.append("<td>" + v + "</td>");
                    });
                    tbody.append("<td><a class=\"nav-link\" href=\"#module=objects&action=get&sha256=" + sha256 + "\">View</a></td>");
                    tbody.append('</tr>');


                    /*$.each(object.submissions, function(k, v){
                        $('#objects-get-form p[name="submissions"]').append('<a href="#module=submissions&action=get&id='+v+'">'+v+'</a>');
                    });*/
                })
                // TODO ANALYZOVANIE OBJEKTOV
                /*analyze_modal_build([[r.result.Object.sha256, r.result.Object.obj_name, r.result.Object.source]]);*/
            }
        },
    });
}

function download_object(sha256){
    $.ajax({
        type: 'POST',
        url: current_env.get('api_url'),
        processData: false,
        contentType: 'application/json',
        data: JSON.stringify({
            module: "objects",
            action: "download",
            parameters: {"sha256":sha256}
        }),
        success: function(r) {
            if(r.error !== ""){
                $.growl.warning({ title: "An error occured!", message: r.error, size: 'large' });
            } else {
                var link = document.createElement("a");
                link.download = sha256;
                link.target = "_blank";

                link.href = 'data:application/octet-stream;base64,'+r.result.Bytes;
                document.body.appendChild(link);
                link.click();

                document.body.removeChild(link);
            }
        },
    });
}

// load and append analyze modal
$.ajax({
    url: "modules/objects/analyze-modal.html",
    success: function (data) { $('#content').append(data); },
    dataType: 'html'
});

// get object data
get_all_objects();

// bind functions to buttons
$('#objects-get-btn-download').on('click', function(){
    download_object(current_env.get('url_hash').get('sha256'));
});

$('#objects-get-btn-results').on('click', function(){
    window.location.href = '#module=results&action=search&filter_sha256='+current_env.get('url_hash').get('sha256');
});
