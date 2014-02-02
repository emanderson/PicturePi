goog.provide('picturepi.pictures');
goog.provide('picturepi.pictures.Picture');

goog.require('goog.dom');
goog.require('goog.ui.FormPost');
goog.require('goog.events');

picturepi.pictures.makePictures = function(data, pictureContainer) {
    var pictures = [];
    for (var i = 0; i < data.length; i++) {
	var picture =
	    new picturepi.pictures.Picture(data[i].fullURL, data[i].previewURL, data[i].fileName, data[i].parentDir, pictureContainer);
	pictures.push(picture);
	picture.makePictureDom();
    }
    goog.events.listen(goog.dom.getElement("zipSelectedLink"), goog.events.EventType.CLICK, picturepi.pictures.logSelected, false, pictures);    
    return pictures;
};

picturepi.pictures.logSelected = function(event) {
    event.preventDefault();
    var selectedFiles = [];
    for (var i = 0; i < this.length; i++) {
	if (this[i].selected) {
	    selectedFiles.push(this[i].fileName);
	}
    }
    console.log(selectedFiles);
    if (selectedFiles.length > 0) {
	var formData = {
	    'path': this[0].parentDir,
	    'selectedFiles': selectedFiles
	};
	var post = new goog.ui.FormPost();
	post.post(formData, '/zipSelected');
    }
}

picturepi.pictures.Picture = function(fullURL, previewURL, fileName, parentDir, pictureContainer) {
    this.fullURL = fullURL;
    this.previewURL = previewURL;
    this.fileName = fileName;
    this.parentDir = parentDir;
    this.selected = false;
    this.parent = pictureContainer;
};

picturepi.pictures.Picture.prototype.toggleSelected = function() {
    if (this.selected) {
	this.selected = false;
	goog.dom.classes.remove(this.divElement, 'thumbnailSelected');
    } else {
	this.selected = true;
	goog.dom.classes.add(this.divElement, 'thumbnailSelected');
    }
};

picturepi.pictures.Picture.prototype.downloadFile = function(event) {
    event.stopPropagation();
};

picturepi.pictures.Picture.prototype.makePictureDom = function() {
    this.imageElement = goog.dom.createDom('img', {'src': this.previewURL, 'width': 160, 'height': 120});
    this.divElement = goog.dom.createDom('div', {'class': 'thumbnail'}, this.imageElement);
    this.fileLinkElement = goog.dom.createDom('a', {'href': this.fullURL}, this.fileName);
    this.fileNameElement = goog.dom.createDom('span', {'class': 'fileName'}, this.fileLinkElement);
    goog.dom.appendChild(this.divElement, this.fileNameElement);
    goog.dom.appendChild(this.parent, this.divElement);
    goog.events.listen(this.divElement, goog.events.EventType.CLICK, this.toggleSelected, false, this);
    goog.events.listen(this.fileLinkElement, goog.events.EventType.CLICK, this.downloadFile, false, this);
};
