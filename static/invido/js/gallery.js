
export default () => {
    let _dataImages = {}
    let _mapImg = new Map()
    let _mapStcks = new Map()
    let _idArray = []
    let _currentImg = {}
    let _image = null
    let _imageOverlay = null
    let _nextBtn = null
    let _prevBtn = null

    function resetStrc() {
        _dataImages = {}
        _mapStcks = new Map()
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
                    const data_imgs = data.images
                    for (let key in data_imgs) {
                        if (data_imgs.hasOwnProperty(key)) {
                            _dataImages = data_imgs[key].sort((a, b) => a.id.localeCompare(b.id))
                            let index = 0
                            _mapImg = new Map()
                            _idArray = []
                            _dataImages.forEach(item => {
                                _mapImg.set(item.id, { name: item.name, redux: item.redux, caption: item.caption, ix: index })
                                _idArray.push({id: item.id, dataid: key})
                                index += 1
                            })
                            _mapStcks.set(key, {idarray: _idArray, mapimg: _mapImg})
                        }
                    }

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
        displayImage(id, dataid) {
            console.log('display image dataid, id ', dataid, id)
            let stackItem = _mapStcks.get(dataid)
            if (!stackItem){
                return
            }
            _mapImg = stackItem.mapimg
            _idArray = stackItem.idarray

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
                const ele = _idArray[index + 1]
                this.displayImage(ele.id, ele.dataid)
            }
        },
        prevImage() {
            const index = _currentImg.ix
            console.log('prev image of', index)
            if (index > 0) {
                const ele = _idArray[index - 1]
                this.displayImage(ele.id, ele.dataid)
            }
        }
    }
}