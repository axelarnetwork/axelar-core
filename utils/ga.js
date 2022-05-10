// log the pageview with their URL
export const pageview = url => {
  if (window) {
    window.gtag("config", "G-81ZT0BK1ZB", {
      page_path: url,
    });
  }
};

// log specific events happening.
export const event = ({ action, params }) => {
  if (window) {
    window.gtag("event", action, params);
  }
};