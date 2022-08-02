export default function Custom404() {
    return <div className="grid h-screen text-white bg-black place-items-center">
        <div className="flex flex-col place-items-center">
        <h1 className="m-12 text-3xl">Hmm, it looks like this page doesn’t exist. </h1>
        <div>Our docs change frequently.</div>
        <div>You should be able to find what you’re looking for at <a href="https://docs.axelar.dev">https://docs.axelar.dev</a></div>
        </div>
    </div>
  }