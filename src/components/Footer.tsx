export const Footer = () => {
  return (
    <footer className="p-12 flex items-center justify-center text-gray-500">
      <p className="text-center">
        This site was served using{" "}
        <a
          className="text-primary underline"
          href="https://github.com/phillip-england/xerus"
          target="_blank"
        >
          xerus
        </a>, a serverside framework I personally developed.
      </p>
    </footer>
  );
};
