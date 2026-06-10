// Language toggle for the ToughRADIUS Handbook (mdBook output only).
//
// The handbook keeps English and Chinese chapters as 1:1 mirrors under
// src/en/ and src/zh/ with identical file names. This script adds a
// language switch link to the menu bar of every rendered page:
//
//   - on /en/<page>.html  -> one link "中文"    pointing to /zh/<page>.html
//   - on /zh/<page>.html  -> one link "English" pointing to /en/<page>.html
//   - on root-level pages (introduction, print) -> links to both language
//     entry pages, since those pages have no single counterpart
//
// Paths are rewritten on the last "/en/" or "/zh/" segment only, so the
// mapping works for any hosting base path (custom domain, github.io
// project pages, `mdbook serve`, or file:// previews).
//
// GitBook renders the same sources through its own pipeline and ignores
// this file; GitBook readers use the per-chapter cross-links instead.
(function () {
  "use strict";

  // counterpart returns the toggle target for the given pathname, or null
  // for pages outside the en/zh tree (handled as "neutral" pages).
  function counterpart(pathname) {
    var enMatch = pathname.match(/^(.*)\/en\/([^/]*)$/);
    if (enMatch) {
      return { href: enMatch[1] + "/zh/" + enMatch[2], label: "中文", lang: "zh" };
    }
    var zhMatch = pathname.match(/^(.*)\/zh\/([^/]*)$/);
    if (zhMatch) {
      return { href: zhMatch[1] + "/en/" + zhMatch[2], label: "English", lang: "en" };
    }
    return null;
  }

  function makeLink(href, label, lang, title) {
    var link = document.createElement("a");
    link.className = "lang-toggle";
    link.href = href;
    link.hreflang = lang;
    link.title = title;
    link.setAttribute("aria-label", title);
    link.textContent = label;
    return link;
  }

  function insertToggle() {
    var buttons = document.querySelector("#mdbook-menu-bar .right-buttons");
    if (!buttons || buttons.querySelector(".lang-toggle")) {
      return;
    }
    var target = counterpart(window.location.pathname);
    if (target) {
      buttons.insertBefore(
        makeLink(
          target.href,
          target.label,
          target.lang,
          target.lang === "zh" ? "切换到中文版本" : "Switch to the English version"
        ),
        buttons.firstChild
      );
      return;
    }
    // Neutral root pages (introduction.html, print.html, directory index):
    // offer both language entry points, resolved relative to the page.
    var zh = makeLink("zh/overview.html", "中文", "zh", "中文手册");
    var en = makeLink("en/overview.html", "English", "en", "English handbook");
    buttons.insertBefore(zh, buttons.firstChild);
    buttons.insertBefore(en, zh);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", insertToggle);
  } else {
    insertToggle();
  }
})();
