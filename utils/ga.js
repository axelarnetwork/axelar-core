// log the pageview with their URL
export const pageview = url => {
  if (window) {
    window.gtag("config", process.env.NEXT_PUBLIC_GA_TRACKING_ID, {
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