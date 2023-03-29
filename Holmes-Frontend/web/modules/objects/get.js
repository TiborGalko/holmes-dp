'use strict';

function get_object(sha256){
	$.ajax({
		type: 'POST',
		url: current_env.get('api_url'),
		processData: false,
		contentType: 'application/json',
		data: JSON.stringify({
			module: "objects",
			action: "get",
			parameters: {"sha256":sha256}
		}),
		success: function(r) {
			if(r.error != ""){
				$.growl.warning({ title: "An error occured!", message: r.error, size: 'large' });
			} else {
				$.each(r.result.Object, function(k, v){
					$('#objects-get-form input[name="'+k+'"]').val(v);
				});

				$.each(r.result.Object.submissions, function(k, v){
					$('#objects-get-form p[id="submissions"]').append('<a href="#module=submissions&action=get&id=' + v + '&sha256='+ r.result.Object.sha256 +'">' + v + '</a><br/>');
				});

				analyze_modal_build([[r.result.Object.sha256, r.result.Object.obj_name, r.result.Object.source]]);
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
			if(r.error != ""){
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
get_object(current_env.get('url_hash').get('sha256'));

// bind functions to buttons
$('#objects-get-btn-download').on('click', function(){
	download_object(current_env.get('url_hash').get('sha256'));
});
