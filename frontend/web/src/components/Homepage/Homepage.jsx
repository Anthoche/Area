import React, { useState } from "react";
import "./homepage.css";
import SearchBar from "./SearchBar";
import FilterTag from "./FilterTag";
import ServiceCard from "./ServiceCard";

export default function Homepage() {
  const [activeCard, setActiveCard] = useState(null);
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const items = Array.from({ length: 9 }).map((_, index) => ({
    title: index === 0 ? "Push To Ping" : `Service ${index + 1}`,
    color: getColor(index),
  }));

  const handleCardClick = (item) => {
    setActiveCard(item);
    setIsPanelOpen(true);
  };

  const closePanel = () => {
    setIsPanelOpen(false);
  };

  return (
    <div className={`layout-root ${sidebarOpen ? "sidebar-open" : ""}`}>
      <aside className="sidebar">
        <div className="sidebar-title">Discovery</div>
        <nav className="sidebar-nav">
          <button>test</button>
        </nav>
      </aside>
      
      <div className={`content-container ${isPanelOpen ? "panel-open" : ""}`}>
        <main className="main-content">
          <div className="top-row">
            <button className="menu-btn" onClick={() => setSidebarOpen(!sidebarOpen)}>
              ☰
            </button>
            <h1 className="main-title">My Konect</h1>
            <button className="profile-btn">👤</button>
          </div>
          <SearchBar />

          <div className="tags-row">
            <FilterTag label="test 1" />
            <FilterTag label="test 2" />
            <FilterTag label="test 3" />
            <FilterTag label="test 4" />
            <FilterTag label="test 5" />
          </div>
          <h2 className="section-header">My Konects</h2>
          <div className="services-grid">
            {items.map((item, idx) => (
              <ServiceCard key={idx} title={item.title} color={item.color} onClick={() => handleCardClick(item)}/>
            ))}
          </div>
          <div className="fab" title="Add new item">+</div>
        </main>

        {isPanelOpen && (
          <aside className="right-panel">
            <button className="close-btn" onClick={closePanel}>✖</button>
            <h2>{activeCard?.title}</h2>
            <p>test</p>
          </aside>
        )}
      </div>
    </div>
  );
}

function getColor(i) {
  const colors = ["#00D2FF", "#FF4081", "#FF4081", "#00E676", "#D500F9"];
  return colors[i % colors.length];
}
