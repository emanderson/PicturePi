goog.provide('picturepi.pictures');
goog.provide('picturepi.pictures.Picture');

goog.require('goog.dom');
goog.require('goog.ui.Zippy');

picturepi.pictures.makePictures = function(data, pictureContainer) {
    var pictures = [];
    for (var i = 0; i < data.length; i++) {
	var picture =
	    new picturepi.pictures.Picture(data[i].fullURL, data[i].previewURL, pictureContainer);
	pictures.push(picture);
	picture.makePictureDom();
    }
    return pictures;
};

picturepi.pictures.Picture = function(fullURL, previewURL, pictureContainer) {
    this.fullURL = fullURL;
    this.previewURL = previewURL;
    this.parent = pictureContainer;
};

picturepi.pictures.Picture.prototype.makePictureDom = function() {
    this.imageElement = goog.dom.createDom('img', {'src': this.previewURL})
    this.linkElement = goog.dom.createDom('a', {'href': this.fullURL}, this.imageElement);
    goog.dom.appendChild(this.parent, this.linkElement);
};
