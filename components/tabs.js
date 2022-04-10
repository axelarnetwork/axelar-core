import { useState, useEffect } from "react";
import { useSelector, useDispatch, shallowEqual } from "react-redux";

import { ENVIRONMENT_DATA } from '../reducers/types';

export default ({ tabs, children }) => {
  const dispatch = useDispatch();
  const { environment } = useSelector(state => ({ environment: state.environment }), shallowEqual);
  const { environment_data } = { ...environment };

  const [openTab, setOpenTab] = useState(0);

  useEffect(() => {
    setOpenTab(environment_data && tabs?.filter(tab => !tab.hidden).findIndex(tab => tab?.title?.toLowerCase() === environment_data?.toLowerCase()) > -1 ?
      tabs.filter(tab => !tab.hidden).findIndex(tab => tab.title.toLowerCase() === environment_data.toLowerCase()) : 0
    );
  }, [environment_data]);

  const onClick = (tab, key) => {
    setOpenTab(key);
    dispatch({
      type: ENVIRONMENT_DATA,
      value: tab?.title?.toLowerCase() || key,
    });
  };

  return (
    <div className="flex flex-wrap flex-col w-full tabs mt-4">
      <div className="flex lg:flex-wrap flex-row lg:space-x-2">
        {tabs?.filter(tab => !tab.hidden).map((tab, key) => (
          <div key={key} className="flex-none">
            <button
              onClick={() => onClick(tab, key)}
              className={openTab === key ? 'tab tab-underline tab-active' : 'tab tab-underline'}
              type="button"
            >
              {tab.title}
            </button>
          </div>
        ))}
      </div>
      {tabs?.filter(tab => !tab.hidden).map((tab, key) => (
        <div key={key} className={`tab-content ${openTab !== key ? 'hidden' : 'block'} max-w-full`}>
          {tab.content}
        </div>
      ))}
    </div>
  );
};