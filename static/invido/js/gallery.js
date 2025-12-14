
export default () => {
    let _dataImages = {}
    let _mapImg = new Map()
    let _idArray = []
    let _currentImg = {}
    let _image = null
    let _imageOverlay = null
    let _nextBtn = null
    let _prevBtn = null

    function resetStrc() {
        _dataImages = {}
        _mapImg = new Map()
        _idArray = []
        _currentImg = {}
        _image = null
        _imageOverlay = null
        _nextBtn = null
        _prevBtn = null
    }
    return {
        loadData() {
            console.log('load data')
            resetStrc()
            fetch('photos.json', { cache: 'no-store' })
                .then(response => response.json())
                .then((data) => {
                    //console.log('data from fetch: ', data)
                    _dataImages = data;
                    _dataImages = data.images.sort((a, b) => a.id.localeCompare(b.id))
                    let index = 0
                    _dataImages.forEach(item => {
                        _mapImg.set(item.id, { name: item.name, redux: item.redux, caption: item.caption, ix: index })
                        _idArray.push(item.id)
                        index += 1
                    })
                    //console.log('dataimages: ', _dataImages)
                    //console.log('mapImg: ', _mapImg)
                    _image = document.querySelector('#the-image');
                    _imageOverlay = document.querySelector('#image-wrapper');
                    _prevBtn = document.querySelector('#previous-btn');
                    _nextBtn = document.querySelector('#next-btn');
                    console.log('images data for gallery ok')
                })
                .catch(err => {
                    console.error('error on fetch: ', err)
                });
        },
        displayImage(id) {
            console.log('display image id ', id)
            _currentImg = _mapImg.get(id)
            _image.classList.add('hidden')
            _image.onload = () => { _image.classList.remove('hidden'); };
            _image.src = _currentImg.name
            _image.alt = _currentImg.caption
            //console.log('current image ', _currentImg) 
            const index = _currentImg.ix
            if (index < _idArray.length - 1) {
                _nextBtn.classList.remove('hidden')
            } else {
                _nextBtn.classList.add('hidden')
            }
            if (index > 0) {
                _prevBtn.classList.remove('hidden')
            } else {
                _prevBtn.classList.add('hidden')
            }
            _imageOverlay.classList.remove('gone');
        },
        hideGalleryImage() {
            console.log('hide gallery image')
            _imageOverlay.classList.add('gone');
        },
        nextImage() {
            const index = _currentImg.ix
            console.log('next image of', index)
            if (index < _idArray.length - 1) {
                this.displayImage(_idArray[index + 1])
            }
        },
        prevImage() {
            const index = _currentImg.ix
            console.log('prev image of', index)
            if (index > 0) {
                this.displayImage(_idArray[index - 1])
            }
        }
    }
}