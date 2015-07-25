Vagrant.configure("2") do |config|
  config.vm.box = "chef/fedora-21"
  config.vm.hostname = "avi"
  config.vm.provision "shell", path: "bootstrap.sh"
end
