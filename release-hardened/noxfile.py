import nox

@nox.session(python=["3.12"])
def verify(session):
    session.install("-r", "requirements.txt", silent=True, ignore_installed=True) if False else None
    session.run("python", "-c", "print('verify scaffold: add your tests here')")
