(function () {
	const storageKey = "forum-theme";
	const root = document.documentElement;
	const media = window.matchMedia("(prefers-color-scheme: dark)");

	function getStoredTheme() {
		try {
			const value = window.localStorage.getItem(storageKey);
			return value === "light" || value === "dark" ? value : "";
		} catch {
			return "";
		}
	}

	function setStoredTheme(theme) {
		try {
			window.localStorage.setItem(storageKey, theme);
		} catch {
			return;
		}
	}

	function resolvedTheme() {
		return getStoredTheme() || (media.matches ? "dark" : "light");
	}

	function applyTheme(theme) {
		root.dataset.theme = theme;
	}

	function updateToggle(theme) {
		const button = document.querySelector("[data-theme-toggle]");
		const label = document.querySelector("[data-theme-toggle-label]");
		if (button) {
			button.setAttribute("aria-pressed", theme === "dark" ? "true" : "false");
		}
		if (label) {
			label.textContent = theme === "dark" ? "Dark" : "Light";
		}
	}

	applyTheme(resolvedTheme());

	function initToggle() {
		const button = document.querySelector("[data-theme-toggle]");
		updateToggle(resolvedTheme());
		if (!button) {
			return;
		}
		button.addEventListener("click", function () {
			const nextTheme = root.dataset.theme === "dark" ? "light" : "dark";
			setStoredTheme(nextTheme);
			applyTheme(nextTheme);
			updateToggle(nextTheme);
		});
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", initToggle);
	} else {
		initToggle();
	}

	function syncSystemTheme() {
		if (!getStoredTheme()) {
			const theme = resolvedTheme();
			applyTheme(theme);
			updateToggle(theme);
		}
	}

	if (typeof media.addEventListener === "function") {
		media.addEventListener("change", syncSystemTheme);
	} else if (typeof media.addListener === "function") {
		media.addListener(syncSystemTheme);
	}
})();
