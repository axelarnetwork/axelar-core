import { combineReducers } from "redux";

import environment from "./environment";
import chain_id from "./chain_id";

export default combineReducers({
  environment,
  chain_id,
});