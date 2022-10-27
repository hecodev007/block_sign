pragma solidity ^0.5.0;

import "./IERC20.sol";
import "./SafeMath.sol";
import "./SafeERC20.sol";
import "./Ownable.sol";
import "./Pausable.sol";

contract TokenVesting is Ownable,Pausable {
    using SafeMath for uint256;
    using SafeERC20 for IERC20;

    IERC20 public token;

    uint256 public planCount = 0;
    uint256 public payPool = 0;
    //uint256 public block_timestamp = 0;

    //A token holder's plan
    struct Plan {
        //beneficiary of tokens after they are released
        address beneficiary;

        //Lock start time
        uint256 startTime;

        //Lock deadline
        uint256 locktoTime;

        //Number of installments of release time
        uint256 releaseStages;

        //Release completion time
        uint256 endTime;

        //Allocated token balance
        uint256 totalToken;

        //Current release quantity
        uint256 releasedAmount;

        //Whether the token can be revoked
        bool revocable;

        //Whether the token has been revoked
        bool isRevoked;

        //Remarks for a plan
        string remark;
    }

    //Token holder's plan set
    mapping (address => Plan) public plans;

    event Released(address indexed beneficiary, uint256 amount);
    event Revoked(address indexed beneficiary, uint256 refund);
    event AddPlan(address indexed beneficiary, uint256 startTime, uint256 locktoTime, uint256 releaseStages, uint256 endTime, uint256 totalToken, uint256 releasedAmount, bool revocable, bool isRevoked, string remark);
    event TransferAddPlanFailure(address indexed beneficiary, uint256 startTime, uint256 locktoTime, uint256 releaseStages, uint256 endTime, uint256 totalToken, uint256 releasedAmount, bool revocable, bool isRevoked, string remark);

    /**
     * @param _token ERC20 token which is being vested
     */
    constructor(address _token) public {
        token = IERC20(_token);
    }

    //function setBlockTimestamp(uint256 _timestamp) public {
    //   block_timestamp = _timestamp;
    //}

    //    /**
    //     * @dev Check if the payment amount of the contract is sufficient
    //     */
    //    modifier checkPayPool(uint256 _totalToken) {
    //        require(token.balanceOf(address(this)) >= payPool.add(_totalToken));
    //        payPool = payPool.add(_totalToken);
    //        _;
    //    }

    /**
     * @dev Check if the plan exists
     */
    modifier whenPlanExist(address _beneficiary) {
        require(_beneficiary != address(0));
        require(plans[_beneficiary].beneficiary != address(0));
        _;
    }

    /**
     * @dev Add a token holder's plan
     */
    function addPlan(address _beneficiary, uint256 _startTime, uint256 _locktoTime, uint256 _releaseStages, uint256 _endTime, uint256 _totalToken, bool _revocable, string memory _remark) internal  whenNotPaused(){
        require(_beneficiary != address(0));
        require(plans[_beneficiary].beneficiary == address(0));

        require(_startTime > 0 && _locktoTime > 0 && _releaseStages > 0 && _totalToken > 0);
        require(_locktoTime > block.timestamp && _locktoTime >= _startTime  && _endTime > _locktoTime);

        require(token.balanceOf(address(this)) >= payPool.add(_totalToken));
        payPool = payPool.add(_totalToken);

        plans[_beneficiary] = Plan(_beneficiary, _startTime, _locktoTime, _releaseStages, _endTime, _totalToken, 0, _revocable, false, _remark);
        planCount = planCount.add(1);
        emit AddPlan(_beneficiary, _startTime, _locktoTime, _releaseStages, _endTime, _totalToken, 0, _revocable, false, _remark);
    }


    /**
 * @dev Add a token transfer for holder's plan
 */
    function transferAndAddPlan(address _beneficiary, uint256 _startTime, uint256 _locktoTime, uint256 _releaseStages, uint256 _endTime, uint256 _totalToken, bool _revocable, string memory _remark) public onlyOwner whenNotPaused(){
        require(_beneficiary != address(0));
        require(plans[_beneficiary].beneficiary == address(0));

        require(_startTime > 0 && _locktoTime > 0 && _releaseStages > 0 && _totalToken > 0);
        require(_locktoTime > block.timestamp && _locktoTime >= _startTime  && _endTime > _locktoTime);

        require(_locktoTime > block.timestamp && _locktoTime >= _startTime  && _endTime > _locktoTime);

        require(token.balanceOf(msg.sender) >= _totalToken);

        if(token.transfer(address(this), _totalToken)) {
            //  plans[_beneficiary] = Plan(_beneficiary, _startTime, _locktoTime, _releaseStages, _endTime, _totalToken, 0, _revocable, false, _remark);
            //  planCount = planCount.add(1);
            //  payPool = payPool.add(_totalToken);
            //  emit AddPlan(_beneficiary, _startTime, _locktoTime, _releaseStages, _endTime, _totalToken, 0, _revocable, false, _remark);
            addPlan(_beneficiary, _startTime, _locktoTime, _releaseStages, _endTime, _totalToken,_revocable, _remark);
        } else {
            emit TransferAddPlanFailure(_beneficiary, _startTime, _locktoTime, _releaseStages, _endTime, _totalToken, 0, _revocable, false, _remark);

        }
    }

    /**
    * @notice Transfers vested tokens to beneficiary.
    */
    function release(address _beneficiary) public whenPlanExist(_beneficiary) whenNotPaused() {

        require(!plans[_beneficiary].isRevoked);

        uint256 unreleased = releasableAmount(_beneficiary);

        if(unreleased > 0 && unreleased <= plans[_beneficiary].totalToken) {
            plans[_beneficiary].releasedAmount = plans[_beneficiary].releasedAmount.add(unreleased);
            payPool = payPool.sub(unreleased);
            token.safeTransfer(_beneficiary, unreleased);
            emit Released(_beneficiary, unreleased);
        }
    }

    /**
     * @dev Calculates the amount that has already vested but hasn't been released yet.
     */
    function releasableAmount(address _beneficiary) public view whenPlanExist(_beneficiary)  returns (uint256) {
        return vestedAmount(_beneficiary).sub(plans[_beneficiary].releasedAmount);
    }

    /**
     * @dev Calculates the amount that has already vested.
     */
    function vestedAmount(address _beneficiary) public view whenPlanExist(_beneficiary)  returns (uint256) {

        if (block.timestamp <= plans[_beneficiary].locktoTime) {
            return 0;
        } else if (plans[_beneficiary].isRevoked) {
            return plans[_beneficiary].releasedAmount;
        } else if (block.timestamp > plans[_beneficiary].endTime && plans[_beneficiary].totalToken == plans[_beneficiary].releasedAmount) {
            return plans[_beneficiary].totalToken;
        }

        uint256 totalTime = plans[_beneficiary].endTime.sub(plans[_beneficiary].locktoTime);
        uint256 totalToken = plans[_beneficiary].totalToken;
        uint256 releaseStages = plans[_beneficiary].releaseStages;
        uint256 endTime = block.timestamp > plans[_beneficiary].endTime ? plans[_beneficiary].endTime : block.timestamp;
        uint256 passedTime = endTime.sub(plans[_beneficiary].locktoTime);

        uint256 unitStageTime = totalTime.div(releaseStages);
        uint256 unitToken = totalToken.div(releaseStages);
        uint256 currStage = passedTime.div(unitStageTime);

        uint256 totalBalance = 0;
        if(currStage > 0 && releaseStages == currStage && (totalTime % releaseStages) > 0 && block.timestamp < plans[_beneficiary].endTime) {
            totalBalance = unitToken.mul(releaseStages.sub(1));
        } else if(currStage > 0 && releaseStages == currStage) {
            totalBalance = totalToken;
        } else if(currStage > 0) {
            totalBalance = unitToken.mul(currStage);
        }
        return totalBalance;

    }

    /**
     * @notice Allows the owner to revoke the vesting. Tokens already vested
     * remain in the contract, the rest are returned to the owner.
     * @param _beneficiary address of the beneficiary to whom vested tokens are transferred
     */
    function revoke(address _beneficiary) public onlyOwner whenPlanExist(_beneficiary) whenNotPaused() {

        require(plans[_beneficiary].revocable && !plans[_beneficiary].isRevoked);

        //Transfer the attribution token to the beneficiary before revoking.
        release(_beneficiary);

        uint256 refund = revokeableAmount(_beneficiary);

        plans[_beneficiary].isRevoked = true;
        payPool = payPool.sub(refund);

        token.transfer(owner(), refund);
        emit Revoked(_beneficiary, refund);
    }

    /**
     * @dev Calculates the amount that recoverable token.
     */
    function revokeableAmount(address _beneficiary) public view whenPlanExist(_beneficiary)  returns (uint256) {

        uint256 totalBalance = 0;

        if(plans[_beneficiary].isRevoked) {
            totalBalance = 0;
        } else if (block.timestamp <= plans[_beneficiary].locktoTime) {
            totalBalance = plans[_beneficiary].totalToken;
        } else {
            totalBalance = plans[_beneficiary].totalToken.sub(vestedAmount(_beneficiary));
        }
        return totalBalance;
    }

    /**
     * Current token balance of the contract
     */
    function thisTokenBalance() public view returns (uint256) {
        return token.balanceOf(address(this));
    }

}