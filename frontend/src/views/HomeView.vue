<template>
  <section>
    <div className="container">
      <h1>Dozingo</h1>
      <p>Willkommen auf der Startseite.</p>

      <div className="grid">
        <div className="col-6 sm-12">
          <div className="card">
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
        </div>
        <div className="col-6 sm-12">
          <div className="card">
            <h2>Buttons</h2>
            <a href="" className="btn btn-primary">Primär</a> 
            <a href="" className="btn btn-secondary">Sekundär</a> 
            <a href="" className="btn btn-success">Erfolg</a> 
            <a href="" className="btn btn-warning">Warnung</a> 
            <a href="" className="btn btn-danger">Fehler</a>
          </div>
        </div>
      </div>
    </div>
  </section>
</template>