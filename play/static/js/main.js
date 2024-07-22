var submitForm = function(el) {
	if (!el.action) {
		return false;
	}

	var req = new XMLHttpRequest();
	req.open('POST', el.action);
	req.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');

	req.onreadystatechange = function() {
		if (this.readyState != 4) {
			return;
		}

		if (this.status == 200) {
			document.getElementById('preview-generated').innerHTML = this.responseText;
			document.getElementById('preview-error').innerHTML = '';
		} else {
			document.getElementById('preview-error').innerHTML = this.getResponseHeader('X-Error-Message');
		}
	};

	var data = encodeFormData(new FormData(el)) + '&download';
	req.send(data);

	return true;
};

var encodeFormData = function(data) {
	var encode = function(s) {
		return encodeURIComponent(s).replace(/%20/g, '+');
	};

	var result = '';
	for (var entry of data.entries()) {
		if (typeof entry[1] == 'string') {
			result += (result ? '&' : '') + encode(entry[0]) + '=' + encode(entry[1]);
		}
	}

	return result;
};
