import React from 'react';
import { Route, Routes } from 'react-router-dom';
import ShortyApp from '../pages/Home';
import Shorty from '../pages/Auth';
import Statistics from '../pages/Statistic';

const RedirectLToBackend = () => {
  const path = window.location.pathname;

  const backendPath = path.replace(/^\/l\//, '');
  window.location.href = `https://backend-wandering-dust-8240.fly.dev/${backendPath}`;
  return null;
};

const RoutesConfig = () => {
  return (
    <Routes>
      <Route path="/home" element={<ShortyApp />} />
      <Route path="/login" element={<Shorty />} />
      <Route path="/statistic" element={<Statistics />} />

      <Route path="/l/*" element={<RedirectLToBackend />} />
    </Routes>
  );
};

export default RoutesConfig;
