<template>
  <section>
    <div className="container">
      <p>Overview for all components and their usage.</p>

      <div className="grid">
        <div className="col-6 sm-12">
          <div className="card card-border-yellow">
            <h2>Alerts</h2>
            <div className="alert alert-success">
              Das ist eine Erfolgsmeldung. 
            </div>
            <div className="alert alert-warning">
              Das ist eine Warnmeldung. 
            </div>
            <div className="alert alert-danger">
              Das ist eine Fehlermeldung. 
            </div>
          </div>
          <div className="card card-border-green mt-3">
            <h2>Dropdown</h2>
            <div class="dropdown">
              <button href="" className="btn btn-secondary dropdown-button">Dropdown</button> 
              <ul class="dropdown-menu">
                <li>My Profile</li>
                <li>Logout</li>
              </ul>
            </div>
          </div>
        </div>
        <div className="col-6 sm-12">
          <div className="card card-border-blue">
            <h2>Buttons</h2>
            <a href="" className="btn btn-primary">Primär</a> 
            <a href="" className="btn btn-secondary">Sekundär</a> 
            <a href="" className="btn btn-success">Erfolg</a> 
            <a href="" className="btn btn-warning">Warnung</a> 
            <a href="" className="btn btn-danger">Fehler</a>
          </div>
          <div className="card card-border-blue mt-3">
            <h2>Badges</h2>
            <span class="badge badge-primary">Primary</span> 
            <span class="badge badge-success">Success</span>
            <span class="badge badge-warning">Warning</span>
            <span class="badge badge-danger">Danger</span>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>