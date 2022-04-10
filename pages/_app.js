import { Provider } from "react-redux";
import { useStore } from "../store";

import "../styles/globals.css";
import "../styles/components/tabs.css";
import "../styles/components/cards.css";
import "nextra-theme-docs/style.css";

export default function Nextra({ Component, pageProps }) {
  const store = useStore(pageProps.initialReduxState);
  const getLayout = Component.getLayout || ((page) => page);
  return (
    <Provider store={store}>
      {getLayout(<Component {...pageProps} />)}
    </Provider>
  );
}
