import { useRouter } from "next/router";
import { useEffect } from "react";
import { Provider } from "react-redux";

import { useStore } from "../store";
import * as ga from "../utils/ga";
import "../styles/globals.css";
import "../styles/components/tabs.css";
import "../styles/components/cards.css";
import "nextra-theme-docs/style.css";

export default ({
  Component,
  pageProps,
}) => {
  const router = useRouter();

  const store = useStore(pageProps.initialReduxState);
  const getLayout = Component.getLayout || (page => page);

  useEffect(() => {
    const handleRouteChange = url => ga.pageview(url);
    //When the component is mounted, subscribe to router changes
    //and log those page views
    router.events.on("routeChangeComplete", handleRouteChange);

    // If the component is unmounted, unsubscribe
    // from the event with the `off` method
    return () => router.events.off("routeChangeComplete", handleRouteChange);
  }, [router.events]);

  return (
    <Provider store={store}>
      {getLayout(
        <Component { ...pageProps } />
      )}
    </Provider>
  );
}