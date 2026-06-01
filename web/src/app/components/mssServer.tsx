import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {AgarBatchSelector, NewAgarBatchForm} from "@/app/components/agarBatchClient";
import {AgarBatchData, ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {MssSelector, NewMssForm} from "@/app/components/mssClient";


export function TestMssOk(){
    const a: MssData = {
        _id: "(MSS ID HERE)",
        creationDate: Date.now()-2000,
        species: "(SPECIES NAME)",
        subspecies: "(SUBSPECIES NAME)",
        parent: "(PARENT ID)",
        transfersOut: ["(TRANSFER 1)","(TRANSFER 2)"],
        sale: "(SALE ID)",
        disposed: Date.now()+40000,
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}
export interface MssData {
    _id: string
    creationDate: number
    species: string
    subspecies?: string
    parent?: string
    transfersOut?: string[]
    sale?: string
    disposed?: number
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
}

export function MssSelectorCloseable(sp: SelectorProps<MssData>) {
    const doSel = (val?: MssData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<MssData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        msgTxt: ChannelTextNewAgarBatch, // TODO: change/get rid of
        closeTxt: "Close MSS List",
        createTxt: "Create MSS",
        lowercase: "mss",
        creatorInPage: sp.creatorInPage,
        createEndpt: "mss",
        getId: (v: MssData) => v._id,
        createSelector:(selHdl: (onSelect: MssData) => void)=>{
            return <MssSelector allowCreate={sp.allowCreation} doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        createCreator:(selHdl: (onSelect: MssData) => void)=>{
            return <NewMssForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}