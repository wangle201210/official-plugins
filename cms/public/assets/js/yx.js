// Tab 切换
function switchTab(index) {
    var btns = document.querySelectorAll('.tab-btn');
    var panes = document.querySelectorAll('.tab-pane');
    btns.forEach(function(btn) { btn.classList.remove('active'); });
    panes.forEach(function(pane) { pane.classList.remove('active'); });
    btns[index].classList.add('active');
    panes[index].classList.add('active');
}

// 轮播图
var slides = document.querySelectorAll('.carousel-slide');
var dots = document.querySelectorAll('.carousel-dots span');

if (slides.length) {
    var currentSlide = 0;

    function goToSlide(index) {
        slides[currentSlide].classList.remove('active');
        dots[currentSlide].classList.remove('active');
        currentSlide = (index + slides.length) % slides.length;
        slides[currentSlide].classList.add('active');
        dots[currentSlide].classList.add('active');
    }

    function changeSlide(dir) {
        goToSlide(currentSlide + dir);
    }

    var carouselTimer = setInterval(function() { changeSlide(1); }, 4000);
    var carouselEl = document.getElementById('carousel');
    carouselEl.addEventListener('mouseenter', function() { clearInterval(carouselTimer); });
    carouselEl.addEventListener('mouseleave', function() {
        carouselTimer = setInterval(function() { changeSlide(1); }, 4000);
    });
}

// 专家照片 & 荣誉资质自动滚动
function autoScroll(trackId, speed) {
    var track = document.getElementById(trackId);
    if (!track) return;
    var container = track.parentElement;
    var pos = 0;
    var paused = false;

    container.addEventListener('mouseenter', function() { paused = true; });
    container.addEventListener('mouseleave', function() { paused = false; });

    setInterval(function() {
        if (paused) return;
        var maxScroll = track.scrollWidth - container.clientWidth;
        if (maxScroll <= 0) return;
        pos += speed;
        if (pos > maxScroll) pos = 0;
        track.style.transform = 'translateX(-' + pos + 'px)';
    }, 30);
}

autoScroll('expertTrack', 0.5);
autoScroll('honorTrack', 0.6);

// 友情链接：点击展开/收起，互斥
document.addEventListener('click', function(e) {
    var title = e.target.closest('.link-group-title');
    if (title) {
        var group = title.parentElement;
        var wasOpen = group.classList.contains('open');
        document.querySelectorAll('.link-group.open').forEach(function(g) {
            g.classList.remove('open');
        });
        if (!wasOpen) group.classList.add('open');
        return;
    }
    // 点击外部收起
    document.querySelectorAll('.link-group.open').forEach(function(g) {
        g.classList.remove('open');
    });
});
