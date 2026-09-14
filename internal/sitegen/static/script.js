// Progressive enhancement only: the page is fully usable without this file
// (every .ics link and its absolute-URL text are already real HTML).

(function () {
  // Local/dev builds with no known base URL render the relative href as
  // placeholder text; fill in the real absolute URL once we know the page's
  // own location.
  document.querySelectorAll('code.ics-url[data-needs-abs-url]').forEach(function (code) {
    var href = code.getAttribute('data-href');
    if (!href) return;
    code.textContent = new URL(href, document.baseURI).href;
  });

  if (!navigator.clipboard) return;

  document.querySelectorAll('.copy-btn').forEach(function (btn) {
    btn.hidden = false;
    btn.addEventListener('click', function () {
      var card = btn.closest('.calendar-card');
      var code = card && card.querySelector('.ics-url');
      if (!code) return;
      var label = btn.querySelector('.copy-label');
      navigator.clipboard.writeText(code.textContent).then(function () {
        var original = label.textContent;
        label.textContent = btn.dataset.copiedLabel || original;
        setTimeout(function () {
          label.textContent = original;
        }, 1500);
      });
    });
  });
})();
