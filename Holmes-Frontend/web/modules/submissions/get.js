'use strict';

function get_submission(id, sha256){
	$.ajax({
		type: 'POST',
		url: current_env.get('api_url'),
		processData: false,
		contentType: 'application/json',
		data: JSON.stringify({
			module: "submissions",
			action: "get",
			parameters: {"id":id, "sha256": sha256}
		}),
		success: function(r) {
			if(r.error != ""){
				$.growl.warning({ title: "An error occured!", message: r.error, size: 'large' });
			} else {
				$.each(r.result.Submission, function(k, v){
					$('#submissions-get-form input[name="'+k+'"]').val(v);
				});
			}
		},
	});
}

get_submission(current_env.get('url_hash').get('id'), current_env.get('url_hash').get('sha256'));
