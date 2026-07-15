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

	function countCharacters(value) {
		return Array.from(value).length;
	}

	function initCommentCounters() {
		document.querySelectorAll(".comment-form").forEach(function (form) {
			const textarea = form.querySelector("[data-comment-textarea]");
			const counter = form.querySelector("[data-comment-counter]");
			const submit = form.querySelector('button[type="submit"]');
			if (!textarea || !counter || !submit) {
				return;
			}

			const limit = Number.parseInt(textarea.dataset.commentLimit, 10);
			const warningThreshold = Number.parseInt(textarea.dataset.commentWarningThreshold, 10);
			if (!Number.isFinite(limit) || !Number.isFinite(warningThreshold)) {
				return;
			}

			function updateCounter() {
				const count = countCharacters(textarea.value);
				const overLimit = count > limit;
				const showWarning = count >= warningThreshold;

				counter.textContent = showWarning ? count + " / " + limit + " characters" : "";
				counter.classList.toggle("is-over-limit", overLimit);
				submit.disabled = overLimit;
			}

			textarea.addEventListener("input", updateCounter);
			updateCounter();
		});
	}

	if (document.readyState === "loading") {
		document.addEventListener("DOMContentLoaded", initCommentCounters);
	} else {
		initCommentCounters();
	}
})();
