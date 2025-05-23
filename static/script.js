document.addEventListener('DOMContentLoaded', function() {
    document.getElementById('searchForm').addEventListener('submit', function(e) {
        e.preventDefault();
        const query = document.getElementById('query').value;
        fetch('/search?q=' + encodeURIComponent(query))
            .then(resp => resp.json())
            .then(showResults)
            .catch(err => console.error(err));
    });
});

function showResults(data) {
    render('googleResults', data.google);
    render('bingResults', data.bing);
    render('baiduResults', data.baidu);
}

function render(id, results) {
    const container = document.getElementById(id);
    container.innerHTML = '';
    results.forEach(r => {
        const div = document.createElement('div');
        const a = document.createElement('a');
        a.href = r.url;
        a.textContent = r.title;
        a.target = '_blank';
        div.appendChild(a);
        container.appendChild(div);
    });
}
