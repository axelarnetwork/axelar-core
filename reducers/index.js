import { combineReducers } from "redux";

import environment from "./environment";
import chain_id from "./chain-id";
import spending_time from "./spending-time";

export default combineReducers({
  environment,
  chain_id,
  spending_time,
});