// App.jsx hoặc App.tsx
import React from 'react';
import { BrowserRouter as Router } from 'react-router-dom';
import RoutesConfig from './routes/route';





const App = () => {
  return (
    <Router>

        <RoutesConfig />
 
    </Router>
  );
};

export default App;
