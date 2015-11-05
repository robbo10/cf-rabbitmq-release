package integration_test

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Upgrading RabbitMQ", func() {
	execBin := func(args ...string) *gexec.Session {
		cmd := exec.Command(binPath, args...)
		session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		return session
	}

	var (
		args    []string
		session *gexec.Session

		tmpDir  string
		tmpFile string
	)

	itExitsWithZero := func() {
		It("exits with a zero exit code", func() {
			Eventually(session).Should(gexec.Exit(0))
		})
	}

	itExitsWithNonZero := func() {
		It("exits with a non-zero exit code", func() {
			Expect(session.ExitCode()).NotTo(BeZero())
		})
	}

	JustBeforeEach(func() {
		session = execBin(args...)
	})

	BeforeEach(func() {
		args = []string{
			"-rabbitmqctl-path", "/idontexist/rabbitmqctl",
			"-node", "node",
			"-new-rabbitmq-version", "0.0.1",
		}

		var err error
		tmpDir, err = ioutil.TempDir("", "rabbitmq-upgrade-tests")
		Expect(err).NotTo(HaveOccurred())
		tmpFile = filepath.Join(tmpDir, "dummy.txt")

		os.Setenv("TEST_OUTPUT_FILE", tmpFile)
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tmpDir)).To(Succeed())
	})

	Context("When there is no new version of rabbit", func() {
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			args = []string{
				"-rabbitmqctl-path", filepath.Join(cwd, "test-assets", "rabbitmqctl-dummy.sh"),
				"-node", "my-node",
				"-new-rabbitmq-version", "3.4.3.1",
			}
		})

		itExitsWithZero()

		It("doesn't call stop app", func() {
			_, err := os.Stat(tmpFile)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("when there is a new patch version of rabbit", func() {
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())
			args = []string{
				"-rabbitmqctl-path", filepath.Join(cwd, "test-assets", "rabbitmqctl-dummy.sh"),
				"-node", "my-node",
				"-new-rabbitmq-version", "3.4.4.1",
			}
		})

		itExitsWithZero()

		It("doesn't call stop app", func() {
			_, err := os.Stat(tmpFile)
			Expect(os.IsNotExist(err)).To(BeTrue())
		})
	})

	Context("when there is a new minor version of rabbit", func() {
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			args = []string{
				"-rabbitmqctl-path", filepath.Join(cwd, "test-assets", "rabbitmqctl-dummy.sh"),
				"-node", "my-node",
				"-new-rabbitmq-version", "3.5.6",
			}
		})

		itExitsWithZero()

		It("calls stop app", func() {
			Eventually(session).Should(gexec.Exit())

			contents, err := ioutil.ReadFile(tmpFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal([]byte("-n my-node\n")))
		})
	})

	Context("when there is a new major version of rabbit", func() {
		BeforeEach(func() {
			cwd, err := os.Getwd()
			Expect(err).NotTo(HaveOccurred())

			args = []string{
				"-rabbitmqctl-path", filepath.Join(cwd, "test-assets", "rabbitmqctl-dummy.sh"),
				"-node", "my-node",
				"-new-rabbitmq-version", "4.4.0.0",
			}
		})

		itExitsWithZero()

		It("calls stop app", func() {
			Eventually(session).Should(gexec.Exit())

			contents, err := ioutil.ReadFile(tmpFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(contents).To(Equal([]byte("-n my-node\n")))
		})
	})

	Context("When the rabbitmqctl-path is not provided", func() {
		BeforeEach(func() {
			args = []string{
				"-node", "node",
				"-new-rabbitmq-version", "0.0.1",
			}
		})

		itExitsWithNonZero()

		It("provides a meaningful error", func() {
			Eventually(session.Err).Should(gbytes.Say("Missing -rabbitmqctl-path flag"))
		})
	})

	Context("When the node is not provided", func() {
		BeforeEach(func() {
			args = []string{
				"-rabbitmqctl-path", "/tmp/rabbitmqctl",
				"-new-rabbitmq-version", "0.0.1",
			}
		})

		itExitsWithNonZero()

		It("provides a meaningful error", func() {
			Eventually(session.Err).Should(gbytes.Say("Missing -node flag"))
		})
	})

	Context("When the rabbitmq-version is not provided", func() {
		BeforeEach(func() {
			args = []string{
				"-rabbitmqctl-path", "/tmp/rabbitmqctl",
				"-node", "node",
			}
		})

		itExitsWithNonZero()

		It("provides a meaningful error", func() {
			Eventually(session.Err).Should(gbytes.Say("Missing -new-rabbitmq-version flag"))
		})
	})
})
