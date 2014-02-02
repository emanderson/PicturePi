goog.provide('picturepi.pictures');
goog.provide('picturepi.pictures.Picture');

goog.require('goog.dom');
goog.require('goog.ui.Zippy');

picturepi.pictures.makePictures = function(data, pictureContainer) {
    var pictures = [];
    for (var i = 0; i < data.length; i++) {
	var picture =
	    new picturepi.pictures.Picture(data[i].fullURL, data[i].previewURL, data[i].fileName, pictureContainer);
	pictures.push(picture);
	picture.makePictureDom();
    }
    return pictures;
};

picturepi.pictures.Picture = function(fullURL, previewURL, fileName, pictureContainer) {
    this.fullURL = fullURL;
    this.previewURL = previewURL;
    this.fileName = fileName;
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

picturepi.pictures.Picture.prototype.makePictureDom = function() {
    this.imageElement = goog.dom.createDom('img', {'src': this.previewURL, 'width': 160, 'height': 120});
    this.divElement = goog.dom.createDom('div', {'class': 'thumbnail'}, this.imageElement);
    goog.dom.appendChild(this.parent, this.divElement);
    goog.events.listen(this.divElement, goog.events.EventType.CLICK, this.toggleSelected, false, this);
};
