import { SPENDING_TIME_SECONDS } from "./types";

export default (
  state = {
    [`${SPENDING_TIME_SECONDS}`]: 30,
  },
  action,
) => {
  switch (action.type) {
    case SPENDING_TIME_SECONDS:
      return {
        ...state,
        [`${SPENDING_TIME_SECONDS}`]: action.value,
      };
    default:
      return state;
  }
};