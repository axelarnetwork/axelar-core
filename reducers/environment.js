import { ENVIRONMENT_DATA } from "./types";

export default function environment(
  state = {
    [`${ENVIRONMENT_DATA}`]: 'mainnet',
  },
  action
) {
  switch (action.type) {
    case ENVIRONMENT_DATA:
      return {
        ...state,
        [`${ENVIRONMENT_DATA}`]: action.value,
      };
    default:
      return state;
  }
};