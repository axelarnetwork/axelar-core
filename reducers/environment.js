import { ENVIRONMENT_ID } from "./types";

export default (
  state = {
    [`${ENVIRONMENT_ID}`]: "mainnet",
  },
  action,
) => {
  switch (action.type) {
    case ENVIRONMENT_ID:
      return {
        ...state,
        [`${ENVIRONMENT_ID}`]: action.value,
      };
    default:
      return state;
  }
};