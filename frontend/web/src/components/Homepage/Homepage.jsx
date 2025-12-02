import React from "react";
import "./homepage.css";
import SearchBar from "./SearchBar";
import FilterTag from "./FilterTag";
import ServiceCard from "./ServiceCard";

export default function Homepage() {
  const items = Array.from({ length: 6 }).map((_, index) => ({
    title: index === 0 ? "Push To Ping" : `Service ${index + 1}`,
    color: getColor(index),
    icons: index === 0 ? ["💻"] : [],
  }));

  return (
    <div className="layout-root">
      <aside className="sidebar">
        <div className="sidebar-title">Discovery</div>
        <nav className="sidebar-nav">
          <button>Trending</button>
          <button>New Releases</button>
          <button>Marketplace</button>
        </nav>

        <div className="sidebar-title">My Library</div>
        <nav className="sidebar-nav">
          <button>My Konects</button>
          <button>My Favorites</button>
          <button>Shared</button>
        </nav>
      </aside>

      <main className="main-content">
        <div className="top-row">
          <h1 className="main-title">My Konect</h1>
          <button className="profile-btn">👤</button>
        </div>

        <SearchBar />

        <div className="tags-row">
          <FilterTag label="Label" />
          <FilterTag label="Label" />
          <FilterTag label="Label" />
          <FilterTag label="Label" />
          <FilterTag label="Label" />
        </div>

        <h2 className="section-header">Frame</h2>

        <div className="services-grid">
          {items.map((item, idx) => (
            <ServiceCard
              key={idx}
              title={item.title}
              color={item.color}
              icons={item.icons}
            />
          ))}
        </div>
      </main>

      <aside className="widget-preview">
        <div className="widget-card">
          Raccourci 2
        </div>
        <div className="toggle-container">
          <div className="toggle-switch"></div>
        </div>
      </aside>
    </div>
  );
}

function getColor(i) {
  const colors = ["#00D2FF", "#FF4081", "#FF4081", "#00E676", "#D500F9"];
  return colors[i % colors.length];
}
