// routes.jsx
import React from 'react';
import { Route, Routes } from 'react-router-dom';
import ShortyApp from '../pages/Home';
import Shorty from '../pages/Auth';
import Statistics from '../pages/Statistic';



const RoutesConfig = () => {
  return (
    <Routes>
      <Route path="/home" element={<ShortyApp/>} />
     <Route path="/login" element={<Shorty/>} />
     <Route path="/statistic" element={<Statistics/>} />
    </Routes>
  );
};

export default RoutesConfig;